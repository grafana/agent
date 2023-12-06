import { FC, useEffect, useState } from 'react';

import Table from './Table';
import { ComponentDetail } from './types';

import styles from '../component/ComponentView.module.css';

interface XRayProps {
  component: ComponentDetail;
}

// Summary of a group of series
interface SeriesSummary {
  name: string;
  series_count: number;
  data_point_count_total: number;
}

const renderTableData = () => {
  return [1, 2, 3].map((n, index) => {
    return (
      <tr>
        <td className={styles.nameColumn}>aaa</td>
        <td>
          <pre className={styles.pre}>12</pre>
        </td>
        <td>
          <pre className={styles.pre}>456</pre>
        </td>
      </tr>
    );
  });
};

export const XRayView: FC<XRayProps> = (props) => {
  const [topMetrics, setTopMetrics] = useState<SeriesSummary[] | undefined>(undefined);
  const [topJobs, setTopJobs] = useState<SeriesSummary[] | undefined>(undefined);
  useEffect(function () {
    const worker = async () => {
      // Request is relative to the <base> tag inside of <head>.
      const resp = await fetch(`./api/v0/component/${props.component.localID}/summary?label=__name__`, {
        cache: 'no-cache',
        credentials: 'same-origin',
      });
      const data: any = await resp.json();
      const summs: SeriesSummary[] = [];
      Object.keys(data).forEach(function (key, index) {
        const summ = data[key];
        console.log();
        const ss: SeriesSummary = {
          name: summ.Labels['__name__'],
          series_count: summ.series_count,
          data_point_count_total: summ.data_point_count_total,
        };
        summs.push(ss);
      });
      setTopMetrics(summs);
    };
    worker().catch(console.error);
    const jobworker = async () => {
      // Request is relative to the <base> tag inside of <head>.
      const resp = await fetch(`./api/v0/component/${props.component.localID}/summary?label=job`, {
        cache: 'no-cache',
        credentials: 'same-origin',
      });
      const data: any = await resp.json();
      const summs: SeriesSummary[] = [];
      Object.keys(data).forEach(function (key, index) {
        const summ = data[key];
        console.log();
        const ss: SeriesSummary = {
          name: summ.Labels['__name__'],
          series_count: summ.series_count,
          data_point_count_total: summ.data_point_count_total,
        };
        summs.push(ss);
      });
      setTopJobs(summs);
    };
    jobworker().catch(console.error);
  }, []);
  return (
    <>
      <h2>Metric Analysis</h2>
      <div className={styles.sectionContent}>
        <h3>Top Metrics by Series Count</h3>
        <div className={styles.list}>
          <Table tableHeaders={['Name', 'Series', 'Samples']} renderTableData={renderTableData} style={{ width: '210px' }} />
        </div>
        <h3>Top Jobs by Series Count</h3>
        <div className={styles.list}>
          <Table tableHeaders={['Name', 'Series', 'Samples']} renderTableData={renderTableData} style={{ width: '210px' }} />
        </div>
      </div>
    </>
  );
};
