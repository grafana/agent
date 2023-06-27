import { faCubes } from '@fortawesome/free-solid-svg-icons';

import PeerList from '../features/clustering/PeerList';
import Page from '../features/layout/Page';
import { usePeerInfo } from '../hooks/peerInfo';

function PageClusteringPeers() {
  const peers = usePeerInfo();

  return (
    <Page name="Clustering" desc="List of clustering peers." icon={faCubes}>
      <PeerList peers={peers} />
    </Page>
  );
}

export default PageClusteringPeers;
