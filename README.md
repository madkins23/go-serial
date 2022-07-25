# go-serial

## Problem Description

Go makes it hard to deserialize an object with fields that are interfaces.
This includes map and array values as well as structs.

Serialization of an interface is pretty simple.
The object that fills the interface has a type to guide the process.

Deserialization of generalized Go interfaces is problematic.
Unmarshaling a given type is generally done using reflection,
since the type is present and can be used to guide the process.
An interface may be filled with instances of any type that implements the interface,
so the decoder can't know what type to generate to fill the interface.

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
There is no extra information of this type in JSON or YAML.

It is possible to override the marshaling code for types
that implement specific interfaces that need to be serialized.
Naive implementations of `MarshalXXXX()` to wrap such output  with
an extra level structure to provide the type name may result in stack overflow.
This will be true if the object is wrapped and then the wrapper serialized,
which will then invoke the `MarshalXXXX()` method again within the wrapper.

#### No Index of Type Names

Any serialized type name will be a string.
Go has no support for looking up or creating a `Type` object by type name.
The type names are (or at least the index to them is) removed at compile time.

This is where online solutions generally use a `switch` statement over the type name.
When adding a new type implementing the interface it is necessary
to add a new `case` to the `switch` statement.
This may be seen as a maintenance issue as the new type will
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

### Wrap Interface Values During Serialization

Override the appropriate `MarshalXXXX()` method for any type
containing an interface field.
Generate 'wrapper' map around each object that fills an interface
with a type name field and a contents field for the original object.

Support for wrappers is provided by this project.
Usage is demonstrated in test files.

### Provide a Type Name Index

Types that implement an interface must be registered.
This allows them to be instantiated by type name.

The [`go-type/reg`](https://github.com/madkins23/go-type) package
provides the type name index.
In addition to providing the type name index,
`go-type/reg.Registry` places the burden for registration
on the type that implements the interface,
not on the code that uses that type in an interface.

This seems like a more maintainable approach.
On the negative side, `go-type/reg` uses reflection
to identify types and generate new type instances.
A `switch` solution avoids the use of reflection and may be more performant.

### Unwrap Interface Values During Deserialization

Override the appropriate `UnmarshalXXXX()` method for any type
containing an interface field.
For each interface field in that type:

* Parse the 'wrapper' map to get the type name.
* Use the global `go-type/reg.Registration` object to instantiate the type.
* Parse the 'wrapper' map for the serialized item and unmarshal into the item.

Support for wrappers is provided by this project.
Usage is demonstrated in test files.

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
results in a lot of extra data conversion or duplicate parsing.

The current solution is different for the various serializers.
On the plus side the code is much simpler and
results in less data conversions.
On the minus side it won't be possible to
(for example) just unplug JSON and plug in YAML without a lot of work,
but then, is it _ever_ simple to change data platforms?

Anyone wishing to trace the evolution of the code:
* it started in [`go-utils/typeutils`](https://github.com/madkins23/go-utils),
* was moved into [`go-type`](https://github.com/madkins23/go-type) (`reg` and other),
* and then the serialization bits ended up here.

It will be necessary to delve back into the history of those projects.
Good luck.  ;-)

## Supported Formats

This package supports several serialization formats:

* JSON
* YAML

Ideally BSON would be included but that work has not been done.

## Caveats

1. There is no anchor mechanism at the current time so serializing data
   with repeated references to the same object will deserialize into
   multiple copies of that object.

2. This code _may_ work with non-`struct` objects that implement an
   interface but no testing has been done thus far.

See the [source](https://github.com/madkins23/go-serial) or
[godoc](https://godoc.org/github.com/madkins23/go-serial) for documentation.