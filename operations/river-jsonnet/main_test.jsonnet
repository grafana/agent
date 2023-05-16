local river = import './main.libsonnet';

// The expectations below have to have sorted fields, since Jsonnet won't give
// you the fields back in the original order.

local tests = [
  {
    name: 'Attributes',
    input: {
      array_attr: ['Hello', 50, false],
      bool_attr: true,
      number_attr: 1234,
      object_attr: { a: 5, b: 6 },
      string_attr: 'Hello, world!',
    },
    expect: |||
      array_attr = ["Hello", 50, false]
      bool_attr = true
      number_attr = 1234
      object_attr = {
        "a" = 5,
        "b" = 6,
      }
      string_attr = "Hello, world!"
    |||,
  },
  {
    name: 'Exprs',
    input: {
      expr_attr: river.expr('prometheus.remote_write.default.receiver'),
    },
    expect: |||
      expr_attr = prometheus.remote_write.default.receiver
    |||,
  },
  {
    name: 'Blocks',
    input: {
      [river.block('labeled_block', 'foobar')]: {
        attr_1: 15,
        attr_2: 30,
      },
      [river.block('unlabeled_block')]: {
        attr_1: 15,
        attr_2: 30,
      },
    },
    expect: |||
      labeled_block "foobar" {
        attr_1 = 15
        attr_2 = 30
      }
      unlabeled_block {
        attr_1 = 15
        attr_2 = 30
      }
    |||,
  },
  {
    name: 'Ordered blocks',
    input: {
      [river.block('labeled_block', 'foobar', index=1)]: {
        attr_1: 15,
        attr_2: 30,
      },
      [river.block('unlabeled_block', index=0)]: {
        attr_1: 15,
        attr_2: 30,
      },
    },
    expect: |||
      unlabeled_block {
        attr_1 = 15
        attr_2 = 30
      }
      labeled_block "foobar" {
        attr_1 = 15
        attr_2 = 30
      }
    |||,
  },
  {
    name: 'Nested blocks',
    input: {
      [river.block('outer.block')]: {
        attr_1: 15,
        attr_2: 30,
        [river.block('inner.block')]: {
          attr_3: 45,
          attr_4: 60,
        },
      },
    },
    expect: |||
      outer.block {
        attr_1 = 15
        attr_2 = 30
        inner.block {
          attr_3 = 45
          attr_4 = 60
        }
      }
    |||,
  },
  {
    name: 'Complex example',
    input: {
      attr_1: 'Hello, world!',
      [river.block('some_block', 'foobar')]: {
        attr_1: [0, 1, 2, 3],
        attr_2: { first_name: 'John', last_name: 'Smith' },
        expr: river.expr('env("HOME")'),
      },
    },
    expect: |||
      attr_1 = "Hello, world!"
      some_block "foobar" {
        attr_1 = [0, 1, 2, 3]
        attr_2 = {
          "first_name" = "John",
          "last_name" = "Smith",
        }
        expr = env("HOME")
      }
    |||,
  },
  {
    name: 'List of blocks',
    input: {
      attr_1: 'Hello, world!',

      [river.block('outer_block')]: {
        attr_1: 53,
        [river.block('inner_block', 'labeled')]: [
          { bool: true },
          { bool: false },
        ],
        [river.block('inner_block', 'other_label')]: [
          { bool: true },
          { bool: false },
        ],
      },
    },
    expect: |||
      attr_1 = "Hello, world!"
      outer_block {
        attr_1 = 53
        inner_block "labeled" {
          bool = true
        }
        inner_block "labeled" {
          bool = false
        }
        inner_block "other_label" {
          bool = true
        }
        inner_block "other_label" {
          bool = false
        }
      }
    |||,
  },
  {
    name: 'Indented literals',
    input: {
      attr_1: river.expr('concat([%s])' % river.manifestRiverValue({ hello: 'world' })),
    },
    expect: |||
      attr_1 = concat([{
        "hello" = "world",
      }])
    |||,
  },
  {
    name: 'Pruned expressions',
    input: std.prune({
      expr: river.expr('env("HOME")'),
    }),
    expect: |||
      expr = env("HOME")
    |||,
  },
];

std.map(function(test) (
  assert river.manifestRiver(test.input) == test.expect : (
    |||
      %s FAILED

      EXPECT
      ======
      %s

      ACTUAL
      ======
      %s
    ||| % [test.name, test.expect, river.manifestRiver(test.input)]
  );
  '%s: PASS' % test.name
), tests)
