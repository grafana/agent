---
aliases:
- ../../configuration-language/expressions/operators/
title: Operators
weight: 300
---

# Operators
River uses a set of operators that most should be familiar with. All operations
follow the standard [PEMDAS](https://en.wikipedia.org/wiki/Order_of_operations)
rule for operator precedence.

## Arithmetic operators

Operator | Description
-------- | -----------
`+`      | Adds two numbers.
`-`      | Subtracts two numbers.
`*`      | Multiplies two numbers.
`/`      | Divides two numbers.
`%`      | Computes the remainder after dividing two numbers.
`^`      | Raises the number to the specified power.

The `+` operator can also be used for string concatenation.

## Comparison operators

Operator | Description
-------- | -----------
`==`     | `true` when two values are equal.
`!=`     | `true` when two values are not equal.
`<`      | `true` when the left value is less than the right value.
`<=`     | `true` when the left value is less than or equal to the right value.
`>`      | `true` when the left value is greater than the right value.
`>=`     | `true` when the left value is greater or equal to the right value.

The equality operators `==` and `!=` can be applied to any operands.

On the other hand, for the ordering operators `<` `<=` `>` and `>=` the two
operands must both be _orderable_ and of the same type. The result of the
comparisons are defined as follows:

* Boolean values are equal if they are either both true or both false.
* Numerical (integer and floating-point) values are orderable, in the usual
  way.
* String values are orderable lexically byte-wise.
* Objects are equal if all their fields are equal.
* Array values are equal if their corresponding elements are equal.

## Logical operators

Operator | Description
-------- | -----------
`&&`     | `true` when the both left _and_ right value are `true`.
`\|\|`     | `true` when the either left _or_ right value are `true`.
`!`      | Negates a boolean value.

Logical operators apply to boolean values and yield a boolean result.

## Assignment operator
River uses `=` as its assignment operator.

An assignment statement may only assign a single value.
In assignments, each value must be _assignable_ to the attribute or object key
to which it is being assigned.

* The `null` value can be assigned to any attribute.
* Numerical, string, boolean, array, function, capsule and object types are
  assignable to attributes of the corresponding type.
* Numbers can be assigned to string attributes with an implicit conversion.
* Strings can be assigned to numerical attributes, provided that they represent
  a number.
* Blocks are not assignable.

## Brackets

Brackets | Description
-------- | -----------
`{ }`    | Defines blocks and objects.
`( )`    | Groups and prioritizes expressions.
`[ ]`    | Defines arrays.

In the following example we can see the use of curly braces and square brackets
to define an object and an array.
```river
obj = { app = "agent", namespace = "dev" }
arr = [1, true, 7 * (1+1), 3]
```

## Access operators

Operator | Description
-------- | -----------
`[ ]`    | Access a member of an array or object.
`.`      | Access a named member of an object or an exported field of a component.

River's access operators support accessing of arbitrarily nested values.
Square brackets can be used to access zero-indexed array indices as well as
object fields by enclosing the field name in double quotes.
The dot operator can be used to access both object fields (without double
quotes) and component exports.
```river
obj["app"]
arr[1]

obj.app
local.file.token.content
```
