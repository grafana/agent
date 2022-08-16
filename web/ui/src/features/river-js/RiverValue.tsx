import { FC, ReactElement, Fragment } from 'react';
import { ObjectField, Value, ValueType } from './types';
import styles from './RiverValue.module.css';

export interface RiverValueProps {
  value: Value;
}

/**
 * RiverValue emits a paragraph which represents a River value.
 */
export const RiverValue: FC<RiverValueProps> = (props) => {
  return (
    <p className={styles.value}>
      <ValueRenderer value={props.value} indentLevel={0} />
    </p>
  );
};

type valueRendererProps = RiverValueProps & {
  indentLevel: number;
};

const ValueRenderer: FC<valueRendererProps> = (props) => {
  const value = props.value;

  switch (value.type) {
    case ValueType.NULL:
      return <span className={styles.literal}>null</span>;

    case ValueType.NUMBER:
      return <span className={styles.literal}>{value.value.toString()}</span>;

    case ValueType.STRING:
      // TODO(rfratto): make sure that line breaks get replaced with escaped
      // characters.
      return <span className={styles.string}>"{value.value}"</span>;

    case ValueType.BOOL:
      if (value.value) {
        return <span className={styles.literal}>true</span>;
      }
      return <span className={styles.literal}>false</span>;

    case ValueType.ARRAY:
      return (
        <>
          <span>[</span>
          {value.value.map((element, idx) => {
            return (
              <Fragment key={idx.toString()}>
                <ValueRenderer value={element} indentLevel={props.indentLevel} />
                {idx + 1 < value.value.length ? <span>, </span> : null}
              </Fragment>
            );
          })}
          <span>]</span>
        </>
      );

    case ValueType.OBJECT:
      if (value.value.length === 0) {
        // No elements; return `{}` without any line breaks.
        return (
          <>
            <span>&#123;</span>
            <span>&#125;</span>
          </>
        );
      }

      const partitions = partitionFields(value.value);

      // TODO(rfratto): make sure that keys which are not valid identifiers
      // are wrapped in quotes.
      return (
        <>
          <span>&#123;</span>
          <br />
          {partitions.map((partition) => {
            // Find the maximum field length across all fields in this
            // partition.
            const keyLength = partitionKeyLength(partition);

            return partition.map((element, index) => {
              return (
                <Fragment key={index.toString()}>
                  {getLinePrefix(props.indentLevel + 1)}
                  <span>{partitionKey(element, keyLength)} = </span>
                  <ValueRenderer value={element.value} indentLevel={props.indentLevel + 1} />
                  <span>,</span>
                  <br />
                </Fragment>
              );
            });
          })}
          {getLinePrefix(props.indentLevel)}
          <span>&#125;</span>
        </>
      );

    case ValueType.FUNCTION:
      return <span className={styles.special}>{value.value}</span>;

    case ValueType.CAPSULE:
      return <span className={styles.special}>{value.value}</span>;
  }
};

/**
 * partitionFields partitions fields in an object by fields which should have
 * their equal signs aligned.
 *
 * A field which crosses multiple lines (i.e., recursively contains an object
 * with more than one element) will cause a partition break, placing subsequent
 * fields in another partition.
 */
function partitionFields(fields: ObjectField[]): ObjectField[][] {
  const partitions = [];

  let currentPartition: ObjectField[] = [];
  fields.forEach((field) => {
    currentPartition.push(field);

    if (multilinedValue(field.value)) {
      // Fields which cross multiple lines cause a partition break.
      partitions.push(currentPartition);
      currentPartition = [];
    }
  });

  if (currentPartition.length !== 0) {
    partitions.push(currentPartition);
  }

  return partitions;
}

/** multilinedValue returns true if value recrusively crosses multiple lines. */
function multilinedValue(value: Value): boolean {
  switch (value.type) {
    case ValueType.OBJECT:
      // River objects cross more than one line whenever there is at least one
      // element.
      return value.value.length > 0;

    case ValueType.ARRAY:
      // River arrays cross more than one line if any of their elements cross
      // more than one line.
      return value.value.some((v) => multilinedValue(v));
  }

  // Other values never cross line barriers.
  return false;
}

/**
 * partitionKeyLength returns the length of keys within the partition. The
 * length is determined by the longest field name in the partition.
 */
function partitionKeyLength(partition: ObjectField[]): number {
  let keyLength = 0;

  partition.forEach((f) => {
    const fieldLength = f.key.length;
    if (fieldLength > keyLength) {
      keyLength = fieldLength;
    }
  });

  return keyLength;
}

/**
 * partitionKey returns the text to use to display a key for a field within a
 * partition.
 */
function partitionKey(field: ObjectField, keyLength: number): string {
  if (field.key.length < keyLength) {
    return field.key + ' '.repeat(keyLength - field.key.length);
  }
  return field.key;
}

function getLinePrefix(indentLevel: number): ReactElement | null {
  if (indentLevel === 0) {
    return null;
  }
  return <span>{'\t'.repeat(indentLevel)}</span>;
}
