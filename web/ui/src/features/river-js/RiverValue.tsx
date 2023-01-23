import React, { FC, Fragment } from 'react';

import { ObjectField, Value, ValueType } from './types';

import styles from './RiverValue.module.css';

export interface RiverValueProps {
  value: Value;
  name?: string;
  nthChild?: number;
}

// TO track the length of the object key value pair
const DIVISION_LENGTH = 20;

/**
 * RiverValue emits a paragraph which represents a River value.
 */
export const RiverValue: FC<RiverValueProps> = ({ value, name, nthChild }) => {
  return (
    <div className={styles.value}>
      <ValueRenderer value={value} nthChild={nthChild} name={name} />
    </div>
  );
};

const ValueRenderer: FC<RiverValueProps> = ({ value, name, nthChild }) => {
  const backgroundColor = nthChild && nthChild % 2 === 1 ? '#f4f5f5' : 'white';

  /**
   * Renderer for the river format for target and label object
   */
  const renderGrid = (partition: ObjectField[]) => {
    // const gridTemplateColumns = '10% 1% 69%';

    let maxLength = 0;

    partition.forEach(({ key }) => {
      return key.length > maxLength ? (maxLength = key.length) : maxLength;
    });

    const gridTemplateColumns = maxLength < DIVISION_LENGTH ? '10% 1% 89%' : undefined;

    return partition.map(({ key, value }) => {
      return (
        <div key={key} className={styles['grid-layout']} style={{ backgroundColor: backgroundColor, gridTemplateColumns }}>
          <div className={`${styles['grid-item']} ${styles['grid-key']}`}>{key}</div>
          <div className={styles['grid-item']}>=</div>
          <ValueRenderer value={value} name={name} />
        </div>
      );
    });
  };

  switch (value.type) {
    case ValueType.NULL:
      return <span className={styles.literal}>null</span>;

    case ValueType.NUMBER:
      return <span className={styles.literal}>{value.value.toString()}</span>;

    case ValueType.STRING:
      return <span className={styles.string}>"{escapeString(value.value)}"</span>;

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
                <ValueRenderer value={element} name={name} />
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

      return (
        <>
          <span>&#123;</span>
          <br />
          {partitions.map((partition) => {
            return renderGrid(partition);
          })}
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
 * escapeString escapes special characters in a string so they can be printed
 * inside a River string literal.
 */
function escapeString(input: string): string {
  // TODO(rfratto): this should also escape Unicode characters into \u and \U
  // forms.
  return input.replace(/[\b\f\n\r\t\v\0'"\\]/g, (match) => {
    switch (match) {
      case '\b':
        return '\\b';
      case '\f':
        return '\\f';
      case '\n':
        return '\\n';
      case '\r':
        return '\\r';
      case '\t':
        return '\\t';
      case '\v':
        return '\\v';
      case "'":
        return "\\'";
      case '"':
        return '\\"';
      case '\\':
        return '\\\\';
    }
    return '';
  });
}
