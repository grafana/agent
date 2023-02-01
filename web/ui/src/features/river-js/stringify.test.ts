import { riverStringify } from './stringify';
import { ValueType } from './types';

it('should render null properly', () => {
  const result = riverStringify({ type: ValueType.NULL });
  expect(result).toEqual('null');
});

it('should render numbers properly', () => {
  const result = riverStringify({ type: ValueType.NUMBER, value: 12345 });
  expect(result).toEqual('12345');
});

it('should render strings properly', () => {
  const result = riverStringify({ type: ValueType.STRING, value: 'Hello, world!' });
  expect(result).toEqual('"Hello, world!"');
});

it('should render true bools properly', () => {
  const result = riverStringify({ type: ValueType.BOOL, value: true });
  expect(result).toEqual('true');
});

it('should render false bools properly', () => {
  const result = riverStringify({ type: ValueType.BOOL, value: false });
  expect(result).toEqual('false');
});

it('should render empty arrays properly', () => {
  const result = riverStringify({ type: ValueType.ARRAY, value: [] });
  expect(result).toEqual('[]');
});

it('should render arrays of basic values properly', () => {
  const result = riverStringify({
    type: ValueType.ARRAY,
    value: [
      { type: ValueType.NULL },
      { type: ValueType.NUMBER, value: 12345 },
      { type: ValueType.BOOL, value: true },
      { type: ValueType.STRING, value: 'fizzbuzz' },
    ],
  });
  expect(result).toEqual('[null, 12345, true, "fizzbuzz"]');
});

it('should render empty objects properly', () => {
  const result = riverStringify({ type: ValueType.OBJECT, value: [] });
  expect(result).toEqual('{}');
});

it('should render non-empty objects properly', () => {
  const result = riverStringify({
    type: ValueType.OBJECT,
    value: [
      { key: 'field_a', value: { type: ValueType.BOOL, value: true } },
      { key: 'field_b', value: { type: ValueType.NUMBER, value: 12345 } },
    ],
  });
  expect(result).toEqual(`{
\tfield_a = true,
\tfield_b = 12345,
}`);
});

it('should render nested objects properly', () => {
  const result = riverStringify({
    type: ValueType.OBJECT,
    value: [
      {
        key: 'nested',
        value: {
          type: ValueType.OBJECT,
          value: [
            { key: 'field_a', value: { type: ValueType.BOOL, value: true } },
            { key: 'field_b', value: { type: ValueType.NUMBER, value: 12345 } },
          ],
        },
      },
    ],
  });

  expect(result).toEqual(`{
\tnested = {
\t\tfield_a = true,
\t\tfield_b = 12345,
\t},
}`);
});

it('should align keys in objects properly', () => {
  const result = riverStringify({
    type: ValueType.OBJECT,
    value: [
      { key: 'reallylongname', value: { type: ValueType.BOOL, value: true } },
      { key: 'shortname', value: { type: ValueType.NUMBER, value: 12345 } },
    ],
  });
  expect(result).toEqual(`{
\treallylongname = true,
\tshortname      = 12345,
}`);
});

test('all-in-one', () => {
  const result = riverStringify({
    type: ValueType.ARRAY,
    value: [
      { type: ValueType.NULL },
      { type: ValueType.NUMBER, value: 12345 },
      {
        type: ValueType.OBJECT,
        value: [
          {
            key: 'reallylongname',
            value: { type: ValueType.NUMBER, value: 12345 },
          },
          {
            key: 'nested',
            value: {
              type: ValueType.OBJECT,
              value: [{ key: 'field_a', value: { type: ValueType.BOOL, value: true } }],
            },
          },
          {
            key: 'shortname',
            value: { type: ValueType.NUMBER, value: 12345 },
          },
        ],
      },
      { type: ValueType.BOOL, value: true },
      { type: ValueType.STRING, value: 'fizzbuzz' },
    ],
  });

  expect(result).toEqual(`[null, 12345, {
\treallylongname = 12345,
\tnested         = {
\t\tfield_a = true,
\t},
\tshortname = 12345,
}, true, "fizzbuzz"]`);
});
