import { FC, Fragment, ReactElement } from 'react';
import { RiverValue } from '../../features/river-js/RiverValue';
import { AttrStmt, Body, StmtType } from '../../features/river-js/types';
import { ComponentDetail, ComponentInfo } from './types';
import styles from './ComponentView.module.css';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faCubes, faLink } from '@fortawesome/free-solid-svg-icons';
import ComponentList from './ComponentList';
import { HealthLabel } from './HealthLabel';
import { RiverBlob } from '../river-js/RiverBlob';

export interface ComponentViewProps {
  component: ComponentDetail;
  info: Record<string, ComponentInfo>;
}

export const ComponentView: FC<ComponentViewProps> = (props) => {
  // TODO(rfratto): expand/collapse icon for sections (treat it like Row in grafana dashboard)

  const inInfo = props.component.inReferences.map((id) => props.info[id]);
  const outInfo = props.component.outReferences.map((id) => props.info[id]);

  const argsPartition = partitionBody(props.component.arguments, 'Arguments');
  const exportsPartition = props.component.exports && partitionBody(props.component.exports, 'Exports');
  const debugPartition = props.component.debugInfo && partitionBody(props.component.debugInfo, 'Debug info');

  function partitionTOC(partition: PartitionedBody): ReactElement {
    return (
      <li>
        <a href={'#' + partition.key.join('-')}>{partition.displayName[partition.displayName.length - 1]}</a>
        {partition.inner.length > 0 && (
          <ul>
            {partition.inner.map((next, idx) => {
              return <Fragment key={idx.toString()}>{partitionTOC(next)}</Fragment>;
            })}
          </ul>
        )}
      </li>
    );
  }

  return (
    <div className={styles.page}>
      <nav>
        <h1>Sections</h1>
        <hr />
        <ul>
          <li>
            <a href={'#' + props.component.id}>{props.component.id}</a>
          </li>
          {props.component.rawConfig && (
            <li>
              <a href="#raw-config">Raw config</a>
            </li>
          )}
          {argsPartition && partitionTOC(argsPartition)}
          {exportsPartition && partitionTOC(exportsPartition)}
          {debugPartition && partitionTOC(debugPartition)}
          {props.component.outReferences.length > 0 && (
            <li>
              <a href="#dependencies">Dependencies</a>
            </li>
          )}
          {props.component.inReferences.length > 0 && (
            <li>
              <a href="#dependants">Dependants</a>
            </li>
          )}
        </ul>
      </nav>

      <main className={styles.content}>
        <h1 id={props.component.id}>
          <span className={styles.icon}>
            <FontAwesomeIcon icon={faCubes} />
          </span>
          {props.component.id}
          &nbsp; {/* space to separate the component name and label so double-click selections work */}
          <span className={styles.healthLabel}>
            <HealthLabel health={props.component.health.type} />
          </span>
        </h1>

        <div className={styles.docsLink}>
          <a href={`https://grafana.com/docs/agent/latest/flow/reference/components/${props.component.name}`}>
            Documentation <FontAwesomeIcon icon={faLink} />
          </a>
        </div>

        {props.component.health.message && (
          <blockquote>
            <h1>
              Latest health message{' '}
              {props.component.health.updateTime && (
                <span className={styles.updateTime}>({props.component.health.updateTime})</span>
              )}
            </h1>
            <p>{props.component.health.message}</p>
          </blockquote>
        )}

        {props.component.rawConfig && (
          <section id="raw-config">
            <h2>Raw config</h2>
            <div className={styles.sectionContent}>
              <RiverBlob>{props.component.rawConfig}</RiverBlob>
            </div>
          </section>
        )}

        <ComponentBody partition={argsPartition} />
        {exportsPartition && <ComponentBody partition={exportsPartition} />}
        {debugPartition && <ComponentBody partition={debugPartition} />}

        {props.component.outReferences.length > 0 && (
          <section id="dependencies">
            <h2>Dependencies</h2>
            <div className={styles.sectionContent}>
              <ComponentList components={outInfo} />
            </div>
          </section>
        )}

        {props.component.inReferences.length > 0 && (
          <section id="dependants">
            <h2>Dependants</h2>
            <div className={styles.sectionContent}>
              <ComponentList components={inInfo} />
            </div>
          </section>
        )}
      </main>
    </div>
  );
};

/**
 * partitionBody groups a body by attributes and inner blocks, assigning unique
 * keys for each.
 */
function partitionBody(body: Body, rootKey: string): PartitionedBody {
  function impl(body: Body, displayName: string[], keyPath: string[]): PartitionedBody {
    const attrs: AttrStmt[] = [];
    const inner: PartitionedBody[] = [];

    const blocksWithName: Record<string, number> = {};

    body.forEach((stmt) => {
      switch (stmt.type) {
        case StmtType.ATTR:
          attrs.push(stmt);
          break;
        case StmtType.BLOCK:
          const blockName = stmt.label ? `${stmt.name}.${stmt.label}` : stmt.name;

          // Keep track of how many blocks have this name so they can be given unique IDs.
          if (blocksWithName[blockName] === undefined) {
            blocksWithName[blockName] = 0;
          }
          const number = blocksWithName[blockName];
          blocksWithName[blockName]++;

          const key = blockName + `_${number}`;

          inner.push(impl(stmt.body, displayName.concat([blockName]), keyPath.concat([key])));
          break;
      }
    });

    return {
      displayName: displayName,
      key: keyPath,
      attrs: attrs,
      inner: inner,
    };
  }

  return impl(body, [rootKey], [rootKey]);
}

interface PartitionedBody {
  /** key is a list of unique identifiers for this partitioned body. */
  key: string[];
  /** displayName is a list of friendly identifiers for this partitioned body. */
  displayName: string[];

  attrs: AttrStmt[];
  inner: PartitionedBody[];
}

interface ComponentBodyProps {
  partition: PartitionedBody;
}

const ComponentBody: FC<ComponentBodyProps> = ({ partition }) => {
  const sectionClass = partition.key.length === 1 ? '' : styles.nested;

  return (
    <>
      <section id={partition.key.join('-')} className={sectionClass}>
        {
          // If the partition only has 1 key, then make it an h2.
          // Otherwise, make it an h3.
          partition.displayName.length === 1 ? (
            <h2>{partition.displayName}</h2>
          ) : (
            <h3>
              {partition.displayName.map((val, idx) => {
                return (
                  <Fragment key={idx.toString()}>
                    <span>{val}</span>
                    {idx + 1 < partition.key.length && <span> / </span>}
                  </Fragment>
                );
              })}
            </h3>
          )
        }
        <div className={styles.sectionContent}>
          <table>
            <thead>
              <tr>
                <th className={styles.nameColumn}>Name</th>
                <th className={styles.valueColumn}>Value</th>
              </tr>
            </thead>
            <tbody>
              {partition.attrs.map((attr) => {
                return (
                  <tr key={attr.name}>
                    <td className={styles.nameColumn}>{attr.name}</td>
                    <td className={styles.valueColumn}>
                      <RiverValue value={attr.value} />
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </section>
      {partition.inner.map((body) => {
        return <ComponentBody key={body.key.join('.')} partition={body} />;
      })}
    </>
  );
};
