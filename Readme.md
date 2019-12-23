### Code guidelines

1. ```data``` should not import anything they should just contain core definition of things.

### Features

### Why should we do this ?


Posts

### Modules

1. ```worker``` -> Main job processor which asynchronously invokes list fetchers based on their configs.
11. Main source - https://github.com/didil/goblero/blob/80f8e2dd691f93b790df479051ea7f6a3b97bcf0/pkg/blero/dispatch.go#L87

1. ```storage``` -> Core data store. This module should be independent as all other modules would depend on it.

### Generating protobuf definitions

1. Install protobuf - ```brew install protobuf```

Execute ```protoc --go_out=. raker.proto```