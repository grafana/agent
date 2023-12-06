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
    let isCancelled = false;

    const fetchData = async () => {
      if (!enabled) {
        setLoading(false);
        return;
      }

      setLoading(true);

      try {
        const response = await fetch(`./api/v0/web/debugStream/${componentID}`, {
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
            let dataArr = `${prevValue}${decodedChunk}`.split('\n');

            if (dataArr.length > maxLines) {
              const difference = dataArr.length - maxLines;
              dataArr = dataArr.slice(difference, dataArr.length);
            }
            return dataArr.join('\n');
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
  }, [componentID, enabled]);

  return { loading, error };
};
