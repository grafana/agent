import { useEffect, useState } from 'react';

/**
 * useStreaming ...
 */
export const useStreaming = (componentID: string) => {
  const [data, setData] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        const response = await fetch(`./api/v0/web/debugStream/${componentID}`);
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
        setLoading(false);
        setError((error as Error).message);
      }
    };

    fetchData();
  }, [componentID]);

  return { data, loading, error };
};
