import { ChangeEvent, useState } from 'react';
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
  const [sampleProb, setSampleProb] = useState(1);
  const [sliderProb, setSliderProb] = useState(1);
  const { loading, error } = useStreaming(String(componentID), enabled, sampleProb, setData);

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

  function handleSampleChange(e: ChangeEvent<HTMLInputElement>) {
    const sampleValue = parseFloat(e.target.value);
    setSliderProb(sampleValue);
  }

  function handleSampleChangeComplete() {
    setSampleProb(sliderProb);
    if (enabled) {
      setEnabled(false);
      setTimeout(() => setEnabled(true), 200);
    }
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

  const samplingControl = (
    <div className={styles.sliderContainer}>
      <span className={styles.sliderLabel}>{Math.round(sliderProb * 100)}% Sampling</span>
      <input
        className={styles.slider}
        type="range"
        min="0"
        max="1"
        step="0.01"
        value={sliderProb}
        onChange={handleSampleChange}
        onMouseUp={handleSampleChangeComplete}
      />
    </div>
  );

  const controls = (
    <>
      {samplingControl}
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
      <AutoScroll className={styles.autoScroll} height={document.body.scrollHeight - 260}>
        {data.map((msg) => {
          return (
            <div className={styles.logLine} key={msg}>
              {msg}
            </div>
          );
        })}
      </AutoScroll>
    </Page>
  );
}

export default PageStreaming;
