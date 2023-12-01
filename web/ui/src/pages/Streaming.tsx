import { faNetworkWired } from '@fortawesome/free-solid-svg-icons';

import Page from '../features/layout/Page';
import { useStreaming } from '../hooks/stream';

function PageStreaming() {
  const peers = useStreaming();

  //   return peers;
  return (
    <Page name="Clustering" desc="List of clustering peers" icon={faNetworkWired}>
      {peers};
    </Page>
  );
}

export default PageStreaming;
