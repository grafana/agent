import { faCubes, faLink } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { FC, Fragment, ReactElement } from 'react';
import { Link } from 'react-router-dom';

import { RiverValue } from '../../features/river-js/RiverValue';
import { AttrStmt, Body, StmtType } from '../../features/river-js/types';

import ComponentList from './ComponentList';
import { HealthLabel } from './HealthLabel';
import { ComponentDetail, ComponentInfo } from './types';

import styles from './ComponentView.module.css';

export interface ComponentViewProps {
  component: ComponentDetail;
  info: Record<string, ComponentInfo>;
}

export const ComponentView: FC<ComponentViewProps> = (props) => {
  // TODO(rfratto): expand/collapse icon for sections (treat it like Row in grafana dashboard)
  console.log(props.component);

  const referencedBy = props.component.referencedBy.filter((id) => props.info[id] !== undefined).map((id) => props.info[id]);
  const referencesTo = props.component.referencesTo.filter((id) => props.info[id] !== undefined).map((id) => props.info[id]);

  const argsPartition = partitionBody(props.component.arguments, 'Arguments');
  const exportsPartition = props.component.exports && partitionBody(props.component.exports, 'Exports');
  const debugPartition = props.component.debugInfo && partitionBody(props.component.debugInfo, 'Debug info');

  function partitionTOC(partition: PartitionedBody): ReactElement {
    return (
      <li>
        <Link to={'#' + partition.key.join('-')} target="_top">
          {partition.displayName[partition.displayName.length - 1]}
        </Link>
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
            <Link to={'#' + props.component.id} target="_top">
              {props.component.id}
            </Link>
          </li>
          {argsPartition && partitionTOC(argsPartition)}
          {exportsPartition && partitionTOC(exportsPartition)}
          {debugPartition && partitionTOC(debugPartition)}
          {props.component.referencesTo.length > 0 && (
            <li>
              <Link to="#dependencies" target="_top">
                Dependencies
              </Link>
            </li>
          )}
          {props.component.referencedBy.length > 0 && (
            <li>
              <Link to="#dependants" target="_top">
                Dependants
              </Link>
            </li>
          )}
          {props.component.moduleInfo && (
            <li>
              <Link to="#module" target="_top">
                Module components
              </Link>
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
            <HealthLabel health={props.component.health.state} />
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
              {props.component.health.updatedTime && (
                <span className={styles.updateTime}>({props.component.health.updatedTime})</span>
              )}
            </h1>
            <p>{props.component.health.message}</p>
          </blockquote>
        )}

        <ComponentBody partition={argsPartition} />
        {exportsPartition && <ComponentBody partition={exportsPartition} />}
        {debugPartition && <ComponentBody partition={debugPartition} />}

        {props.component.referencesTo.length > 0 && (
          <section id="dependencies">
            <h2>Dependencies</h2>
            <div className={styles.sectionContent}>
              <ComponentList components={referencesTo} parent={props.component.parent} />
            </div>
          </section>
        )}

        {props.component.referencedBy.length > 0 && (
          <section id="dependants">
            <h2>Dependants</h2>
            <div className={styles.sectionContent}>
              <ComponentList components={referencedBy} parent={props.component.parent} />
            </div>
          </section>
        )}

        {props.component.moduleInfo && (
          <section id="module">
            <h2>Module components</h2>
            <div className={styles.sectionContent}>
              <ComponentList
                components={props.component.moduleInfo}
                parent={pathJoin([props.component.parent, props.component.id])}
              />
            </div>
          </section>
        )}
      </main>
    </div>
  );
};

function pathJoin(paths: (string | undefined)[]): string {
  return paths.filter((p) => p && p !== '').join('/');
}

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
          {partition.attrs.length === 0 ? (
            <em className={styles.informative}>(No set attributes in this block)</em>
          ) : (
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
          )}
        </div>
      </section>
      {partition.inner.map((body) => {
        return <ComponentBody key={body.key.join('.')} partition={body} />;
      })}
    </>
  );
};
