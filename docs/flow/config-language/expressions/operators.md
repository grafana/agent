---
aliases:
- /docs/agent/latest/flow/configuration-language/expressions/operators
title: Operators
weight: 300
---

# Operators
River uses a set of operators that most should be familiar with. All operations
follow the standard [PEMDAS](https://en.wikipedia.org/wiki/Order_of_operations)
rule for operator precedence.

## Arithmetic operators
```
+    for addition: 3+5
-    for subtraction: 10-3
*    for multiplication: 13*7
/    for division: 15/3
%    for remainder: 7%2
^    for exponentiation: 2^8
```

The `+` operator can also be used for string concatenation.

## Comparison operators
```
==    equal
!=    not equal
<     less
<=    less or equal
>     greater
>=    greater or equal
```

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
```
&&    conditional AND
||    conditional OR 
!     NOT            
```

Logical operators apply to boolean values and yield a boolean result.

## Assignment operator
River uses `=` as its assignment operator.

An assignment statement may only assign a single value to a single attribute.
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
```
{ }    used for defining blocks and objects
( )    used to group and prioritize expressions
[ ]    used for defining arrays
```

In the following example we can see the use of curly braces and square brackets
to define an object and an array.
```river
obj = { app = "agent", namespace = "dev" }
arr = [1, true, 7 * (1+1), 3]
```

## Access operators
```
[ ]    used for access operations on arrays and objects
 .     used for access operations on objects and components
```

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
