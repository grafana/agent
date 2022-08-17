import { FC, Fragment, ReactElement, useEffect, useRef } from 'react';
import { RiverValue } from '../../features/river-js/RiverValue';
import { AttrStmt, Body, StmtType } from '../../features/river-js/types';
import { ComponentDetail, ComponentInfo } from './types';
import styles from './ComponentView.module.css';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faCubes } from '@fortawesome/free-solid-svg-icons';
import ComponentList from './ComponentList';
import prism from 'prismjs';
import './RiverPrismTheme.css';
import { HealthLabel } from './HealthLabel';

prism.languages.river = {
  blockHeader: {
    pattern: /^\s*[^=]+{/m,
    inside: {
      selector: {
        pattern: /([A-Za-z_][A-Za-z0-9_]*)(.([A-Za-z_][A-Za-z0-9_]*))*/,
      },
      comment: {
        pattern: /\/\/.*|\/\*[\s\S]*?(?:\*\/|$)/,
        greedy: true,
      },
      string: {
        pattern: /(^|[^\\])"(?:\\.|[^\\"\r\n])*"(?!\s*:)/,
        lookbehind: true,
        greedy: true,
      },
    },
  },
  comment: {
    pattern: /\/\/.*|\/\*[\s\S]*?(?:\*\/|$)/,
    greedy: true,
  },
  number: /-?\b\d+(?:\.\d+)?(?:e[+-]?\d+)?\b/i,
  string: {
    pattern: /(^|[^\\])"(?:\\.|[^\\"\r\n])*"(?!\s*:)/,
    lookbehind: true,
    greedy: true,
  },
  boolean: /\b(?:false|true)\b/,
  null: {
    pattern: /\bnull\b/,
    alias: 'keyword',
  },
};

export interface ComponentViewProps {
  component: ComponentDetail;
  info: Record<string, ComponentInfo>;
}

const RiverBlob: FC<{ children: string }> = ({ children }) => {
  const codeRef = useRef<HTMLPreElement>(null);

  useEffect(() => {
    if (codeRef.current == null) {
      return;
    }

    prism.highlightAllUnder(codeRef.current);
  }, []);

  return (
    <pre ref={codeRef} style={{ margin: '0px', fontSize: '14px' }}>
      <code className="language-river">{children}</code>
    </pre>
  );
};

export const ComponentView: FC<ComponentViewProps> = (props) => {
  // TODO(rfratto): health information after h1
  // TODO(rfratto): expand/collapse icon for sections (treat it like Row in grafana dashboard)
  // TODO(rfratto): bring title before section closer inside of it

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
            <a href="#raw-config">Raw config</a>
          </li>
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
      /
      <main className={styles.content}>
        <h1>
          <span className={styles.icon}>
            <FontAwesomeIcon icon={faCubes} />
          </span>
          {props.component.id}
          &nbsp; {/* space to separate the component name and label so double-click selections work */}
          <span className={styles.healthLabel}>
            <HealthLabel health={props.component.health.type} />
          </span>
        </h1>

        <section id="raw-config">
          <h2>Raw config</h2>
          <div className={styles.sectionContent}>
            <RiverBlob>
              {`metrics.scrape "k8s_pods" {
  targets    = discovery.k8s.pods.targets
  forward_to = [metrics.remote_write.default.receiver]

  scrape_config {
    job_name = "default"
  }
}`}
            </RiverBlob>
          </div>
        </section>

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
