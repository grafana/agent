import { useEffect, useState } from 'react';

/**
 * useStreaming ...
 */
export const useStreaming = () => {
  const [data, setData] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        const response = await fetch('./api/v0/web/streamDatas');
        if (!response.ok || !response.body) {
          throw response.statusText;
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
          console.log(`received data ${decodedChunk}`);
        }
      } catch (error) {
        setLoading(false);
        // Handle other errors
      }
    };

    fetchData();
  }, []);

  return (
    <div>
      <div>
        <b>Request Response: {loading && <i>Fetching data...</i>}</b>
        {data}
      </div>
    </div>
  );
};
