# go-serial

## Problem Description

Go makes it hard to deserialize an object with fields that are interfaces.

Serialization of an interface is pretty simple.
The object that fills the interface at the time of serialization
has a type to guide the process.

Deserialization of generalized Go interfaces is problematic.
Unmarshaling a given type is generally done using reflection,
since the type is present and can be used to guide the process.
An interface may be filled with instances of any type that implements the interface,
so the decoder can't know what type to generate to instantiate the interface.

### Manual Solution

It is possible to write custom code to handle this problem.
Generally it involves:

1. unmarshal to `interface{}` producing `map[string]interface{}`
2. use information in the `map`
   (e.g. a `Type` string field generated during serialization)
   to determine the type of the object
3. allocate an object of that type
4. unmarshal from the map into a pointer of the correct type

There are various examples of this available online.
Normally there is a `switch` statement with type names
for all the types that may be serialized.
This may represent a code maintenance issue.

### Technical Challenges

Some obstacles to generalizing serialization of objects that have interface fields.

#### No Type Names in Serialized Formats

Existing serializing code doesn't recognize the need to represent
interface values as being of specific types.
There is no extra information of this type in JSON or YAML output.
It is possible to override the marshaling code for types
that implement specific interfaces that need to be serialized.

Naive implementations of `MarshalXXXX()` to wrap such output  with
an extra level structure to provide the type name may result in stack overflow.
This will be true if the object is wrapped and then the wrapper serialized,
which will then invoke the `MarshalXXXX()` method again within the wrapper.
The workaround for this is to copy the object's field into a struct defined
purely for the purpose and then to serialize _that_ object.

Note that this will require a lot of custom serialization code.
It will also require a lot of custom _de_-serialization code
in order to read the type name and then handle the appropriate type.

#### No Index of Type Names

Any serialized type name will be a string.
Go has no support for looking up or creating a `Type` object by type name.
The type names are (or at least the index to them is) removed at compile time.

This is where online solutions generally use a `switch` statement over the type name.
When adding a new type implementing the interface it is necessary
to add a new `case` to the `switch` statement.
This may be seen as a maintenance issue as the new type declaration will
not necessarily be anywhere near the `switch` statement and
the developer must know about and maintain this 'magic' connection.

#### No Hook to Instantiate Named Types for Interface Values

Serialization packages provide `UnmarshalXXXX()` methods for
customizing deserialization behavior.
Interfaces declare methods but do not define them.

If type `Alpha` has an interface field than
`Alpha` must implement the `UnmarshalXXXX()` method.
This method must handle the unmarshaling of the interface field.

## The `go-serial` Solution

### Provide a Generic Wrapper Object

For a given serialization type (e.g. JSON or YAML) generate a wrapper
type to hold the object and an envelope containing the type name and
the marshaled data for the object.

```
Wrapper
+-----------------------------------+
| Item to be wrapped                |
+-----------------------------------+
| Envelope                          |
| +-------------------------------+ |
| | Type Name                     | |
| +-------------------------------+ |
| | Raw (marshaled) Data for Item | |
| +-------------------------------+ |
+-----------------------------------+
```

Utilize Go generics to specify the interface type to be wrapped.
This may not be absolutely necessary but it seems to simplify
references to the wrapped items which don't have to be
defined as `{}interface` and type converted after deserialization.

#### Serialization

Construct the custom serialization code for the wrapper object.
During marshaling of the wrapper object:

1. acquire the type name of the contained object for the envelope,
2. marshal the object into a "raw" field in the envelope, and
3. pass the envelope to the serialization package for marshaling.

#### Deserialization

Construct the custom deserialization code from the wrapper object.
During unmarshaling of the wrapper object:

1. unmarshal into the wrapper envelope,
2. instantiate the actual object from the type name in the envelope, and
3. use the serialization package to unmarshal the actual object
   from the "raw" field in the envelope.

#### Code Localized to Wrapper

All of this behavior is attached to the generic wrapper object.
Depending on how the wrapper is used this may reduce custom code
to a bare minimum.

### Provide a Type Name Index

Since Go removes type names during compilation it is necessary to provide
an index into which types that implement interfaces can be registered.
This provides:
* a way to get the name of the type when serializing an object and
* a way to generate a new, empty object of that type (via reflection)
  during deserialization.

The [`go-type/reg`](https://github.com/madkins23/go-type) package
provides the type name index.
In addition to providing the type name index,
`go-type/reg.Registry` places the burden for registration
on the type that implements the interface,
not on the code that uses that type in an interface,
breaking the "magic" link between new types and deserialization code.

This seems like a more maintainable approach.
On the negative side, `go-type/reg` uses reflection
to identify types and generate new type instances.
A `switch` solution avoids the use of reflection and is likely more performant,
but there is no support for that solution in `go-serial` at this time.

## Usage

There are two basic ways to use `go-serial` wrapper objects.

### Use the Wrappers Directly

The intention is that data structures containing interface fields
should define those fields as wrapped interfaces.
Thus a `struct` might be:

```
type ZZZ struct {
   name string
   age int
   job *json.Wrapper[Employer]
   pets []*json.Wrapper[Pet]
}
```

This provides the simplest usage with little or no additional serialization code.
The downside of this is that the data for a field is always kept within
a wrapper and must be dereferenced during use.

### Convert to Wrappers During Serialization

When serializing a data structure that contains interface fields,
generate a shadow structure with wrapper fields and copy the data
back and forth as required using custom marshal and unmarshal code.
The shadow structure is basically the structure that would be created
when using the wrappers directly per the previous section.

This avoids the dereferencing of wrappers to get the wrapped items.
The downside is that any data structure with one or more interface fields
must have custom serialization code and a shadow structure.

### Examples

Usage of the above is demonstrated in the various test files.

## Choices

The solution provided by `go-serial` (and `go-type`) is the result of
a lot of trial and error.
Initial solutions attempted to simplify code by
doing a lot of conversion to Go data maps
(nested `map[string]{interface}` and array structure)
and required a lot more overridden methods to configure.
Interim solutions were simpler but ended up with wrapper data
being encoded or decoded twice in the process.

The early problems seem to have been attempts to overly generalize the solution.
Various serializers are different enough that making them all "work the same"
(i.e. by converting all types though the nested `map[string]{interface}` and array structure)
results in a lot of extra data conversion or duplicate parsing.

The current solution is different for the various serializers.
On the plus side the code is much simpler and
results in less data conversions.
On the minus side it is necessary to plan for and use wrapper objects.

Anyone wishing to trace the evolution of the code:
* it started in [`go-utils/typeutils`](https://github.com/madkins23/go-utils),
* was moved into [`go-type`](https://github.com/madkins23/go-type) (`reg` and other),
* and then the serialization bits ended up here.

It will be necessary to delve back into the history of those projects.
Good luck.  ;-)

## Supported Formats

This package currently supports several serialization formats:

* JSON
* YAML

Some thought was given to splitting this library into multiple
libraries, one per serialization format.
Since this library uses [stretchr/testify](https://github.com/stretchr/testify)
which in turn supports various comparisons including YAML it seemed
reasonable to include these two generally useful formats.

In addition it would be nice to include, for example, BSON.
That, however, would compicate the module dependencies and testing and
seems more appropriate to be included in a separate library.

## Caveats

1. There is no anchor mechanism at the current time so serializing data
   with repeated references to the same object will deserialize into
   multiple copies of that object.

2. This code _may_ work with non-`struct` objects that implement an
   interface but no testing has been done thus far.

See the [source](https://github.com/madkins23/go-serial) or
[godoc](https://godoc.org/github.com/madkins23/go-serial) for documentation.
