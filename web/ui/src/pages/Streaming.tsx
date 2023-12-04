import { useParams } from 'react-router-dom';
import { faNetworkWired } from '@fortawesome/free-solid-svg-icons';

import Page from '../features/layout/Page';
import { useStreaming } from '../hooks/stream';

function PageStreaming() {
  const { componentID } = useParams();
  const { data, loading, error } = useStreaming(String(componentID));

  return (
    <Page name="DebugStream" desc="Debug stream of data" icon={faNetworkWired}>
      {loading && <p>Loading...</p>}
      {error && <p>Error: {error}</p>}
      <pre>{data}</pre>
    </Page>
  );
}

export default PageStreaming;
