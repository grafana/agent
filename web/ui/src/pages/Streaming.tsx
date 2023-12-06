import { useState } from 'react';
import { useParams } from 'react-router-dom';
import AutoScroll from '@brianmcallister/react-auto-scroll';
import { faBroom, faDownload, faRoad, faSkull, faStop } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';

import Page from '../features/layout/Page';
import { useStreaming } from '../hooks/stream';

import styles from './Streaming.module.css';

function PageStreaming() {
  const { componentID } = useParams();
  const [enabled, setEnabled] = useState(true);
  const [data, setData] = useState<string[]>([]);
  const { loading, error } = useStreaming(String(componentID), enabled, setData);

  function toggleEnableButton() {
    if (enabled) {
      return (
        <div className={styles.xrayLink}>
          <button className={styles.stopButton} onClick={() => setEnabled(false)}>
            Stop <FontAwesomeIcon icon={faStop} />
          </button>
        </div>
      );
    }
    return (
      <div className={styles.xrayLink}>
        <button className={styles.resumeButton} onClick={() => setEnabled(true)}>
          Resume <FontAwesomeIcon icon={faRoad} />
        </button>
      </div>
    );
  }

  function downloadData() {
    const blob = new Blob([data.join('\n')], { type: 'text/plain' });
    const href = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = href;
    link.download = `${componentID}-debug.txt`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(href);
  }

  const controls = (
    <>
      {toggleEnableButton()}
      <div className={styles.xrayLink}>
        <button className={styles.clearButton} onClick={() => setData([])}>
          Clear <FontAwesomeIcon icon={faBroom} />
        </button>
      </div>
      <div className={styles.xrayLink}>
        <button className={styles.downloadButton} onClick={downloadData}>
          Download <FontAwesomeIcon icon={faDownload} />
        </button>
      </div>
    </>
  );

  return (
    <Page name="Debug with X-Ray" desc="Debug stream of data" icon={faSkull} controls={controls}>
      {loading && <p>Streaming data...</p>}
      {error && <p>Error: {error}</p>}
      <AutoScroll height={document.body.scrollHeight - 260}>
        {data.map((msg) => {
          return <div key={msg}>{msg}</div>;
        })}
      </AutoScroll>
    </Page>
  );
}

export default PageStreaming;
