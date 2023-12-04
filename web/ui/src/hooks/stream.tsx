import { useEffect, useState } from 'react';

export const useStreaming = (componentID: string) => {
  const [data, setData] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

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

        while (true) {
          const { value, done } = await reader.read();
          if (done) {
            setLoading(false);
            break;
          }

          const decodedChunk = decoder.decode(value, { stream: true });
          setData((prevValue) => `${prevValue}${decodedChunk}`);
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

  return { data, loading, error };
};
