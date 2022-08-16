import { FC, Fragment, ReactElement } from 'react';
import { NavLink } from 'react-router-dom';
import { RiverValue } from '../../features/river-js/RiverValue';
import { AttrStmt, Body, StmtType } from '../../features/river-js/types';
import { ComponentDetail } from './types';
import styles from './ComponentView.module.css';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faAngleRight, faCubes } from '@fortawesome/free-solid-svg-icons';

export interface ComponentViewProps {
  component: ComponentDetail;
}

export const ComponentView: FC<ComponentViewProps> = (props) => {
  // TODO(rfratto): health information after h1

  const args = partitionBody(props.component.arguments, 'Arguments');

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
      <div className={styles.content}>
        <h1>
          <span className={styles.icon}>
            <FontAwesomeIcon icon={faCubes} />
          </span>
          <span className={styles.icon}>
            <FontAwesomeIcon icon={faAngleRight} />
          </span>
          {props.component.id}
        </h1>

        <nav>
          <ul>
            {partitionTOC(args)}
            {props.component.exports && <li>Exports</li>}
            {props.component.debugInfo && <li>Debug info</li>}
            {props.component.outReferences.length > 0 && <li>Dependencies</li>}
            {props.component.inReferences.length > 0 && <li>Dependants</li>}
          </ul>
        </nav>

        <ComponentBody partition={args} />

        {props.component.exports && (
          <section>
            <h2>Exports</h2>
            <div className={styles.sectionContent}></div>
          </section>
        )}

        {props.component.debugInfo && (
          <section>
            <h2>Debug info</h2>
            <div className={styles.sectionContent}></div>
          </section>
        )}

        {props.component.outReferences.length > 0 && (
          <section>
            <h2>Dependencies</h2>
            <div className={styles.sectionContent}>
              <ul>
                {props.component.outReferences.map((ref) => {
                  return (
                    <li key={ref}>
                      <NavLink to={'/component/' + ref}>{ref}</NavLink>
                    </li>
                  );
                })}
              </ul>
            </div>
          </section>
        )}

        {props.component.inReferences.length > 0 && (
          <section>
            <h2>Dependants</h2>
            <div className={styles.sectionContent}>
              <ul>
                {props.component.inReferences.map((ref) => {
                  return (
                    <li key={ref}>
                      <NavLink to={'/component/' + ref}>{ref}</NavLink>
                    </li>
                  );
                })}
              </ul>
            </div>
          </section>
        )}
      </div>
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
                    {idx + 1 < partition.key.length && (
                      <span>
                        <FontAwesomeIcon icon={faAngleRight} />
                      </span>
                    )}
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
