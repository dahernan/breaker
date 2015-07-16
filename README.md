Circuit Breaker Implementation in Go 

* This is my second attempt to implement the circuit breaker pattern in Go, the first one was [goHystrix](https://github.com/dahernan/goHystrix)

* This version is much simple and idiomatic

* It doesn't have any external dependency 

* Inspired by Nexflix Hystrix https://github.com/Netflix/Hystrix

How to use
----------

```go
b, err := NewBreaker(OptionsDefaults())

err = b.Call(func() error {
	// do my logic
	return nil
})

if err == ErrBreakerOpen {
	// the circuit is open
}

```

### Default circuit values when you create a breaker
```
// OptionsDefaults
// ErrorsPercentage - 50 - If number_of_errors / total_calls * 100 > 50.0 the circuit will be open
// MinimumNumberOfRequest - if total_calls < 20 the circuit will be close
// NumberOfSecondsToStore - 20 seconds
```
