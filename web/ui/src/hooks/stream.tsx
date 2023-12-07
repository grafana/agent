import { useEffect, useRef, useState } from 'react';

export const useStreaming = (
  componentID: string,
  enabled: boolean,
  sampleProb: number,
  filterValue: string,
  setData: React.Dispatch<React.SetStateAction<string[]>>
) => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const maxLines = 5000;

  const filterValueRef = useRef(filterValue);

  useEffect(() => {
    filterValueRef.current = filterValue;
  }, [filterValue]);

  useEffect(() => {
    const abortController = new AbortController();
    let isCancelled = false;

    const fetchData = async () => {
      if (!enabled) {
        setLoading(false);
        return;
      }

      setLoading(true);

      try {
        const response = await fetch(`./api/v0/web/debugStream/${componentID}?sampleProb=${sampleProb}`, {
          signal: abortController.signal,
        });
        if (!response.ok || !response.body) {
          throw new Error(response.statusText || 'Unknown error');
        }

        const reader = response.body.getReader();
        const decoder = new TextDecoder();

        while (enabled && !isCancelled) {
          const { value, done } = await reader.read();
          if (done) {
            break;
          }

          const decodedChunk = decoder.decode(value, { stream: true });

          setData((prevValue) => {
            const newValue = decodedChunk
              .slice(0, -6)
              .split('|xray|')
              .filter((n) => n.toLowerCase().includes(filterValueRef.current));
            let dataArr = prevValue.concat(newValue);

            if (dataArr.length > maxLines) {
              dataArr = dataArr.slice(-maxLines);
            }
            return dataArr;
          });
        }
      } catch (error) {
        if (!isCancelled && (error as Error).name !== 'AbortError') {
          setError((error as Error).message);
        }
      } finally {
        if (!isCancelled) {
          setLoading(false);
        }
      }
    };

    fetchData();

    return () => {
      isCancelled = true;
      abortController.abort();
    };
  }, [componentID, enabled, sampleProb]);

  return { loading, error };
};
