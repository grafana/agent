import { useState } from 'react';
import { useParams } from 'react-router-dom';
import { faBroom, faRoad, faSkull, faStop } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';

import Page from '../features/layout/Page';
import { useStreaming } from '../hooks/stream';

import styles from './Streaming.module.css';

function PageStreaming() {
  const { componentID } = useParams();
  const [enabled, setEnabled] = useState(true);
  const [data, setData] = useState('');
  const { loading, error } = useStreaming(String(componentID), enabled, setData);

  function toggleEnableButton() {
    if (enabled) {
      return (
        <div className={styles.xrayLink}>
          <a onClick={() => setEnabled(false)}>
            Stop <FontAwesomeIcon icon={faStop} />
          </a>
        </div>
      );
    }
    return (
      <div className={styles.xrayLink}>
        <a onClick={() => setEnabled(true)}>
          Resume <FontAwesomeIcon icon={faRoad} />
        </a>
      </div>
    );
  }

  const controls = (
    <>
      {toggleEnableButton()}
      <div className={styles.xrayLink}>
        <a onClick={() => setData('')}>
          Clear <FontAwesomeIcon icon={faBroom} />
        </a>
      </div>
    </>
  );

  return (
    <Page name="Debug with X-Ray" desc="Debug stream of data" icon={faSkull} controls={controls}>
      {loading && <p>Streaming data...</p>}
      {error && <p>Error: {error}</p>}
      <pre className={styles.streamingData}>{data}</pre>
    </Page>
  );
}

export default PageStreaming;
