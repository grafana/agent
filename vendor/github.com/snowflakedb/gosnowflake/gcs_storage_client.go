// Copyright (c) 2021-2022 Snowflake Computing Inc. All rights reserved.

package gosnowflake

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	gcsMetadataPrefix             = "x-goog-meta-"
	gcsMetadataSfcDigest          = gcsMetadataPrefix + sfcDigest
	gcsMetadataMatdescKey         = gcsMetadataPrefix + "matdesc"
	gcsMetadataEncryptionDataProp = gcsMetadataPrefix + "encryptiondata"
	gcsFileHeaderDigest           = "gcs-file-header-digest"
)

type snowflakeGcsClient struct {
}

type gcsLocation struct {
	bucketName string
	path       string
}

func (util *snowflakeGcsClient) createClient(info *execResponseStageInfo, _ bool) (cloudClient, error) {
	if info.Creds.GcsAccessToken != "" {
		return info.Creds.GcsAccessToken, nil
	}
	return "", nil
}

// cloudUtil implementation
func (util *snowflakeGcsClient) getFileHeader(meta *fileMetadata, filename string) (*fileHeader, error) {
	if meta.resStatus == uploaded || meta.resStatus == downloaded {
		return &fileHeader{
			digest:             meta.gcsFileHeaderDigest,
			contentLength:      meta.gcsFileHeaderContentLength,
			encryptionMetadata: meta.gcsFileHeaderEncryptionMeta,
		}, nil
	}
	if meta.presignedURL != nil {
		meta.resStatus = notFoundFile
	} else {
		URL, err := util.generateFileURL(meta.stageInfo.Location, strings.TrimLeft(filename, "/"))
		if err != nil {
			return nil, err
		}
		accessToken := meta.client.(string)
		gcsHeaders := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}

		req, err := http.NewRequest("HEAD", URL.String(), nil)
		if err != nil {
			return nil, err
		}
		for k, v := range gcsHeaders {
			req.Header.Add(k, v)
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			if resp.StatusCode == 403 || resp.StatusCode == 408 || resp.StatusCode == 429 || resp.StatusCode == 500 || resp.StatusCode == 503 {
				meta.lastError = fmt.Errorf(resp.Status)
				meta.resStatus = needRetry
				return nil, meta.lastError
			}
			if resp.StatusCode == 404 {
				meta.resStatus = notFoundFile
			} else if util.isTokenExpired(resp) {
				meta.lastError = fmt.Errorf(resp.Status)
				meta.resStatus = renewToken
			}
			meta.lastError = fmt.Errorf(resp.Status)
			meta.resStatus = errStatus
			return nil, meta.lastError
		}

		digest := resp.Header.Get(gcsMetadataSfcDigest)
		contentLength, err := strconv.Atoi(resp.Header.Get("content-length"))
		if err != nil {
			return nil, err
		}
		var encryptionMeta *encryptMetadata
		if resp.Header.Get(gcsMetadataEncryptionDataProp) != "" {
			var encryptData *encryptionData
			json.Unmarshal([]byte(resp.Header.Get(gcsMetadataEncryptionDataProp)), encryptData)
			if encryptData != nil {
				encryptionMeta = &encryptMetadata{
					key: encryptData.WrappedContentKey.EncryptionKey,
					iv:  encryptData.ContentEncryptionIV,
				}
				if resp.Header.Get(gcsMetadataMatdescKey) != "" {
					encryptionMeta.matdesc = resp.Header.Get(gcsMetadataMatdescKey)
				}
			}
		}
		meta.resStatus = uploaded
		return &fileHeader{
			digest:             digest,
			contentLength:      int64(contentLength),
			encryptionMetadata: encryptionMeta,
		}, nil
	}
	return nil, nil
}

// cloudUtil implementation
func (util *snowflakeGcsClient) uploadFile(
	dataFile string,
	meta *fileMetadata,
	encryptMeta *encryptMetadata,
	maxConcurrency int,
	multiPartThreshold int64) error {
	uploadURL := meta.presignedURL
	var accessToken string
	var err error

	if uploadURL == nil {
		_, err = util.generateFileURL(meta.stageInfo.Location, strings.TrimLeft(meta.dstFileName, "/"))
		if err != nil {
			return err
		}
		accessToken = meta.client.(string)
	}

	var contentEncoding string
	if meta.dstCompressionType != nil {
		contentEncoding = strings.ToLower(meta.dstCompressionType.name)
	}

	if contentEncoding == "gzip" {
		contentEncoding = ""
	}

	gcsHeaders := make(map[string]string)
	gcsHeaders[httpHeaderContentEncoding] = contentEncoding
	gcsHeaders[gcsMetadataSfcDigest] = meta.sha256Digest
	if accessToken != "" {
		gcsHeaders["Authorization"] = "Bearer " + accessToken
	}

	if encryptMeta != nil {
		encryptData := encryptionData{
			"FullBlob",
			contentKey{
				"symmKey1",
				encryptMeta.key,
				"AES_CBC_256",
			},
			encryptionAgent{
				"1.0",
				"AES_CBC_256",
			},
			encryptMeta.iv,
			keyMetadata{
				"Java 5.3.0",
			},
		}
		b, err := json.Marshal(&encryptData)
		if err != nil {
			return err
		}
		gcsHeaders[gcsMetadataEncryptionDataProp] = string(b)
		gcsHeaders[gcsMetadataMatdescKey] = encryptMeta.matdesc
	}

	var uploadSrc io.Reader
	if meta.srcStream != nil {
		uploadSrc = meta.srcStream
		if meta.realSrcStream != nil {
			uploadSrc = meta.realSrcStream
		}
	} else {
		uploadSrc, err = os.OpenFile(dataFile, os.O_RDONLY, os.ModePerm)
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequest("PUT", uploadURL.String(), uploadSrc)
	if err != nil {
		return err
	}
	for k, v := range gcsHeaders {
		req.Header.Add(k, v)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		meta.lastError = fmt.Errorf(resp.Status)
		if resp.StatusCode == 403 || resp.StatusCode == 408 || resp.StatusCode == 429 || resp.StatusCode == 500 || resp.StatusCode == 503 {
			meta.lastError = fmt.Errorf(resp.Status)
			meta.resStatus = needRetry
		} else if accessToken == "" && resp.StatusCode == 400 && meta.lastError == nil {
			meta.lastError = fmt.Errorf(resp.Status)
			meta.resStatus = renewPresignedURL
		} else if accessToken != "" && util.isTokenExpired(resp) {
			meta.lastError = fmt.Errorf(resp.Status)
			meta.resStatus = renewToken
		}
		return meta.lastError
	}

	if meta.options.putCallback != nil {
		meta.options.putCallback = &snowflakeProgressPercentage{
			filename:        dataFile,
			fileSize:        float64(meta.srcFileSize),
			outputStream:    meta.options.putCallbackOutputStream,
			showProgressBar: meta.options.showProgressBar,
		}
	}

	meta.dstFileSize = meta.uploadSize
	meta.resStatus = uploaded

	meta.gcsFileHeaderDigest = gcsHeaders[gcsFileHeaderDigest]
	meta.gcsFileHeaderContentLength = meta.uploadSize
	if err = json.Unmarshal([]byte(gcsHeaders[gcsMetadataEncryptionDataProp]), &encryptMeta); err != nil {
		return err
	}
	meta.gcsFileHeaderEncryptionMeta = encryptMeta
	return nil
}

// cloudUtil implementation
func (util *snowflakeGcsClient) nativeDownloadFile(
	meta *fileMetadata,
	fullDstFileName string,
	maxConcurrency int64) error {
	downloadURL := meta.presignedURL
	var accessToken string
	var err error
	gcsHeaders := make(map[string]string)

	if downloadURL == nil {
		downloadURL, err = util.generateFileURL(meta.stageInfo.Location, strings.TrimLeft(meta.dstFileName, "/"))
		if err != nil {
			return err
		}
		accessToken = meta.client.(string)
		if accessToken != "" {
			gcsHeaders["Authorization"] = "Bearer " + accessToken
		}
	}

	req, err := http.NewRequest("GET", downloadURL.String(), nil)
	if err != nil {
		return err
	}
	for k, v := range gcsHeaders {
		req.Header.Add(k, v)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		meta.lastError = fmt.Errorf(resp.Status)
		if resp.StatusCode == 403 || resp.StatusCode == 408 || resp.StatusCode == 429 || resp.StatusCode == 500 || resp.StatusCode == 503 {
			meta.lastError = fmt.Errorf(resp.Status)
			meta.resStatus = needRetry
		} else if accessToken == "" && resp.StatusCode == 400 && meta.lastError == nil {
			meta.lastError = fmt.Errorf(resp.Status)
			meta.resStatus = renewPresignedURL
		} else if accessToken != "" && util.isTokenExpired(resp) {
			meta.lastError = fmt.Errorf(resp.Status)
			meta.resStatus = renewToken
		}
		return meta.lastError
	}

	f, err := os.OpenFile(fullDstFileName, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err = io.Copy(f, resp.Body); err != nil {
		return err
	}

	var encryptMeta encryptMetadata
	if resp.Header.Get(gcsMetadataEncryptionDataProp) != "" {
		var encryptData *encryptionData
		if err = json.Unmarshal([]byte(resp.Header.Get(gcsMetadataEncryptionDataProp)), &encryptData); err != nil {
			return err
		}
		if encryptData != nil {
			encryptMeta = encryptMetadata{
				encryptData.WrappedContentKey.EncryptionKey,
				encryptData.ContentEncryptionIV,
				"",
			}
			if key := resp.Header.Get(gcsMetadataMatdescKey); key != "" {
				encryptMeta.matdesc = key
			}
		}
	}

	fi, err := os.Stat(fullDstFileName)
	if err != nil {
		return err
	}
	meta.srcFileSize = fi.Size()
	meta.resStatus = downloaded
	meta.gcsFileHeaderDigest = resp.Header.Get(gcsMetadataSfcDigest)
	meta.gcsFileHeaderContentLength = resp.ContentLength
	meta.gcsFileHeaderEncryptionMeta = &encryptMeta
	return nil
}

func (util *snowflakeGcsClient) extractBucketNameAndPath(location string) *gcsLocation {
	containerName := location
	var path string
	if strings.Contains(location, "/") {
		containerName = location[:strings.Index(location, "/")]
		path = location[strings.Index(location, "/")+1:]
		if path != "" && strings.HasSuffix(location, "/") {
			path += "/"
		}
	}
	return &gcsLocation{containerName, path}
}

func (util *snowflakeGcsClient) generateFileURL(stageLocation string, filename string) (*url.URL, error) {
	gcsLoc := util.extractBucketNameAndPath(stageLocation)
	fullFilePath := gcsLoc.path + filename
	URL, err := url.Parse("https://storage.googleapis.com/" + gcsLoc.bucketName + "/" + url.QueryEscape(fullFilePath))
	if err != nil {
		return nil, err
	}
	return URL, nil
}

func (util *snowflakeGcsClient) isTokenExpired(resp *http.Response) bool {
	return resp.StatusCode == 401
}
