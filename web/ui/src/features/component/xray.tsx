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
  last_value: number;
}

interface LabelSummary {
  name: string;
  count: number;
}

export const XRayView: FC<XRayProps> = (props) => {
  const [topMetrics, setTopMetrics] = useState<SeriesSummary[]>([]);
  const [topByLabel, setTopByLabel] = useState<SeriesSummary[]>([]);
  const [details, setDetails] = useState<SeriesSummary[]>([]);
  const [topLabels, setTopLabels] = useState<LabelSummary[]>([]);
  const [selectedLabel, setSelectedLabel] = useState<string>('job');
  const [detailQuery, setDetailQuery] = useState<string>('__name__=go_gc_duration_seconds');
  const renderTableData = () => {
    return topMetrics?.map((el) => {
      return (
        <tr>
          <td>{el.name}</td>
          <td>{el.series_count}</td>
          <td>{el.data_point_count_total}</td>
        </tr>
      );
    });
  };
  const renderJobTableData = () => {
    return topByLabel?.map((el) => {
      return (
        <tr>
          <td>{el.name}</td>
          <td>{el.series_count}</td>
          <td>{el.data_point_count_total}</td>
          <td>
            <div className={styles.docsLink}>
              <a onClick={() => setDetailQuery(selectedLabel + '=' + el.name)}>Details</a>
            </div>
          </td>
        </tr>
      );
    });
  };
  const renderDetailTableData = () => {
    return details?.map((el) => {
      return (
        <tr>
          <td>{el.name}</td>
          <td>{el.data_point_count_total}</td>
          <td>{el.last_value}</td>
        </tr>
      );
    });
  };
  const renderTopLabelsTableData = () => {
    return topLabels?.map((el) => {
      return (
        <tr>
          <td>{el.name}</td>
          <td>{el.count}</td>
          <td>
            <div className={styles.docsLink}>
              <a onClick={() => setSelectedLabel(el.name)}>Details</a>
            </div>
          </td>
        </tr>
      );
    });
  };
  useEffect(function () {
    const worker = async () => {
      const resp = await fetch(`./api/v0/component/${props.component.localID}/summary?label=__name__`, {
        cache: 'no-cache',
        credentials: 'same-origin',
      });
      const data: any[] = await resp.json();
      const summs: SeriesSummary[] = [];
      data.slice(0, 20).forEach(function (summ) {
        const ss: SeriesSummary = {
          name: summ.Labels['__name__'],
          series_count: summ.series_count,
          data_point_count_total: summ.data_point_count_total,
          last_value: 0,
        };
        summs.push(ss);
      });
      setTopMetrics(summs);
    };
    worker().catch(console.error);
  }, []);
  useEffect(function () {
    const worker = async () => {
      const resp = await fetch(`./api/v0/component/${props.component.localID}/labels`, {
        cache: 'no-cache',
        credentials: 'same-origin',
      });
      const data: LabelSummary[] = await resp.json();
      setTopLabels(data);
    };
    worker().catch(console.error);
  }, []);
  useEffect(
    function () {
      const worker = async () => {
        const resp = await fetch(`./api/v0/component/${props.component.localID}/summary?label=${selectedLabel}`, {
          cache: 'no-cache',
          credentials: 'same-origin',
        });
        const data: any[] = await resp.json();
        const summs: SeriesSummary[] = [];
        data.slice(0, 20).forEach(function (summ) {
          const ss: SeriesSummary = {
            name: summ.Labels[selectedLabel],
            series_count: summ.series_count,
            data_point_count_total: summ.data_point_count_total,
            last_value: 0,
          };
          summs.push(ss);
        });
        setTopByLabel(summs);
      };
      worker().catch(console.error);
    },
    [selectedLabel]
  );
  useEffect(
    function () {
      const worker = async () => {
        const resp = await fetch(`./api/v0/component/${props.component.localID}/details?${detailQuery}`, {
          cache: 'no-cache',
          credentials: 'same-origin',
        });
        const data: any[] = await resp.json();
        console.log(data);
        const summs: SeriesSummary[] = [];
        data.slice(0, 20).forEach(function (summ) {
          const ss: SeriesSummary = {
            name: summ.LabelsStr,
            series_count: 1,
            data_point_count_total: summ.DataPoints,
            last_value: summ.LastValue,
          };
          summs.push(ss);
        });
        setDetails(summs);
      };
      worker().catch(console.error);
    },
    [detailQuery]
  );
  const handleChange = (event: any) => {
    setDetailQuery(event.target.value);
  };
  const handleLabelChange = (event: any) => {
    setSelectedLabel(event.target.value);
  };
  return (
    <>
      <h2>Metric Analysis</h2>
      <div className={styles.sectionContent}>
        <h3>Top Metrics by Series Count</h3>
        <div className={styles.list}>
          <Table tableHeaders={['Name', 'Series', 'Samples']} renderTableData={renderTableData} />
        </div>
        <h3>
          Top Values for{' '}
          <input type="text" value={selectedLabel} style={{ width: '100px' }} onChange={handleLabelChange}></input> Label by
          Series Count
        </h3>
        <div className={styles.list}>
          <Table tableHeaders={['Name', 'Series', 'Samples']} renderTableData={renderJobTableData} />
        </div>
        <h3>Highest Cardinality Labels</h3>
        <div className={styles.list}>
          <Table tableHeaders={['Name', 'Unique Values', '']} renderTableData={renderTopLabelsTableData} />
        </div>
        <h3>
          Details for Series with:{' '}
          <input type="text" value={detailQuery} style={{ width: '500px' }} onChange={handleChange}></input>
        </h3>
        <div className={styles.list}>
          <Table tableHeaders={['Name', 'Samples', 'LastValue']} renderTableData={renderDetailTableData} />
        </div>
      </div>
    </>
  );
};
