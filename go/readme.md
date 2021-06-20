Welcome to the Go solution.

This solution leverages several techniques:

1. marshalling/unmarshalling JSON responses via `encoding/json` package

2. "table-driven tests" inspired by:
    https://nathanleclaire.com/blog/2015/10/10/interfaces-and-composition-for-effective-unit-testing-in-golang/
    and https://dave.cheney.net/2013/06/09/writing-table-driven-tests-in-go
    and https://dave.cheney.net/2019/05/07/prefer-table-driven-tests

3. automating common commands via makefile

4. orchestrating concurrency via WaitGroup and channels

5. composition/embedding: https://golang.org/doc/effective_go#embedding

6. Leveragin a custom interface that allows setting the func that our Mock will run
instead, and using `init() {...}` to inject a mock for testing. Borrowed from here: https://levelup.gitconnected.com/mocking-outbound-http-calls-in-golang-9e5a044c2555 