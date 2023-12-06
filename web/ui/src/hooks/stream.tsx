import { useEffect, useState } from 'react';

export const useStreaming = (
  componentID: string,
  enabled: boolean,
  setData: React.Dispatch<React.SetStateAction<string>>
) => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const maxLines = 50000;

  useEffect(() => {
    const abortController = new AbortController();
    const fetchData = async () => {
      try {
        setLoading(true);
        const response = await fetch(`./api/v0/web/debugStream/${componentID}`, {
          signal: abortController.signal,
        });
        if (!response.ok || !response.body) {
          throw new Error(response.statusText || 'Unknown error');
        }

        const reader = response.body.getReader();
        const decoder = new TextDecoder();

        while (enabled) {
          const { value, done } = await reader.read();
          if (done) {
            setLoading(false);
            break;
          }

          const decodedChunk = decoder.decode(value, { stream: true });

          setData((prevValue) => {
            let dataArr = `${prevValue}${decodedChunk}`.split('\n');

            if (dataArr.length > maxLines) {
              const difference = dataArr.length - maxLines;
              dataArr = dataArr.slice(difference, dataArr.length);
            }
            return dataArr.join('\n');
          });
        }
      } catch (error) {
        if ((error as Error).name !== 'AbortError') {
          setError((error as Error).message);
        }
      } finally {
        setLoading(false);
      }
    };

    fetchData();

    return () => {
      abortController.abort();
    };
  }, [componentID]);

  return { loading, error };
};
