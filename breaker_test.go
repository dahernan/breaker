package breaker

import (
	"errors"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

var errValueForTest = errors.New("This should fail")

func returnAnError() error {
	return errValueForTest
}
func returnNilError() error {
	return nil
}

func OptionsForTest() Options {
	return Options{
		ErrorsPercentage:       50.0,
		MinimumNumberOfRequest: 3,
		NumberOfSecondsToStore: 2,
	}
}

func TestBreakerReturnsNoError(t *testing.T) {
	Convey("Breaker executes the code normally", t, func() {
		b, err := NewBreaker(OptionsDefaults())
		So(err, ShouldBeNil)
		So(b, ShouldNotBeNil)

		err = b.Call(returnNilError)
		So(err, ShouldBeNil)
	})
}

func TestBreakerReturnsAnError(t *testing.T) {
	Convey("Breaker returns an error", t, func() {

		b, err := NewBreaker(OptionsDefaults())
		So(err, ShouldBeNil)
		So(b, ShouldNotBeNil)

		err = b.Call(returnAnError)
		So(err, ShouldEqual, errValueForTest)

	})
}

func TestNewBreakerReturnsErrorWithBadOptions(t *testing.T) {
	Convey("Breaker returns when you created with wrong options", t, func() {
		_, err := NewBreaker(Options{
			ErrorsPercentage:       50.0,
			MinimumNumberOfRequest: 3,
			NumberOfSecondsToStore: 61,
		})
		So(err, ShouldEqual, ErrNumberOfSecondsToStoreOutOfBounds)

	})
}

func TestBreakerReturnsNoErrorAndTheCircuitIsClose(t *testing.T) {
	Convey("Breaker returns no error and the circuit keeps closed", t, func() {
		var err error

		b, err := NewBreaker(OptionsForTest())
		So(err, ShouldBeNil)
		So(b, ShouldNotBeNil)

		err = b.Call(returnNilError)
		So(err, ShouldBeNil)

		err = b.Call(returnNilError)
		So(err, ShouldBeNil)

		err = b.Call(returnNilError)
		So(err, ShouldBeNil)
	})
}

func TestAfter2ErrorsTheCircuitShouldBeClose(t *testing.T) {
	Convey("After 2 errors, the circuit will be close because it does not match the minimal number of request", t, func() {
		var err error

		b, err := NewBreaker(OptionsForTest())
		So(err, ShouldBeNil)
		So(b, ShouldNotBeNil)

		// 1
		err = b.Call(returnAnError)
		So(err, ShouldEqual, errValueForTest)

		// 2
		err = b.Call(returnAnError)
		So(err, ShouldEqual, errValueForTest)

		// 3
		err = b.Call(returnNilError)
		So(err, ShouldBeNil)
	})
}

func TestAfter3ErrorsTheCircuitShouldBeOpen(t *testing.T) {
	Convey("After 3 errors, the circuit is open, the 4th call should return Breaker Open", t, func() {
		var err error
		var state uint32

		b, err := NewBreaker(OptionsForTest())
		So(err, ShouldBeNil)
		So(b, ShouldNotBeNil)

		// initial to an invalid state
		state = 9999

		// 1
		err = b.Call(returnAnError)
		So(err, ShouldEqual, errValueForTest)

		// 2
		err = b.Call(returnAnError)
		So(err, ShouldEqual, errValueForTest)

		// 3
		err = b.Call(returnAnError)
		So(err, ShouldEqual, errValueForTest)

		// listen to state changes
		go func() {
			state = <-b.Changes()
		}()

		// limit reached, circuit Open

		// gives up the execution, to let other goroutines time to run
		time.Sleep(8 * time.Millisecond)
		So(state, ShouldEqual, OpenState)

		// 4 ErrBreakerOpen
		err = b.Call(returnNilError)
		So(err, ShouldEqual, ErrBreakerOpen)

		counts := b.Health()
		So(counts.Failures, ShouldEqual, int64(3))
	})
}

func TestCircuitShouldRecover(t *testing.T) {
	Convey("Given a circuit is open, when the wait period pass, then the circuit it will be close", t, func() {
		// After 3 errors -> Open
		// Wait period    -> Leave the circuit to recover
		// After wait     -> Close again
		var err error
		var state uint32
		// initial to an invalid state
		state = 9999

		b, err := NewBreaker(OptionsForTest())
		So(err, ShouldBeNil)
		So(b, ShouldNotBeNil)

		// 1
		err = b.Call(returnAnError)
		So(err, ShouldEqual, errValueForTest)

		// 2
		err = b.Call(returnAnError)
		So(err, ShouldEqual, errValueForTest)

		// 3
		err = b.Call(returnAnError)
		So(err, ShouldEqual, errValueForTest)

		// listen to state changes
		go func() {
			for {
				state = <-b.Changes()
			}
		}()

		// limit reached, circuit Open
		// gives up the execution, to let other goroutines time to run
		time.Sleep(8 * time.Millisecond)
		So(state, ShouldEqual, OpenState)

		// 4 ErrBreakerOpen
		err = b.Call(returnNilError)
		So(err, ShouldEqual, ErrBreakerOpen)

		// Wait period
		t.Log("Wait period", b.Health())
		time.Sleep(3 * time.Second)
		t.Log("Wait period finished", b.Health())

		err = b.Call(returnNilError)
		So(err, ShouldBeNil)

		time.Sleep(8 * time.Millisecond)
		So(state, ShouldEqual, CloseState)

		err = b.Call(returnNilError)
		So(err, ShouldBeNil)

		counts := b.Health()
		So(counts.Failures, ShouldEqual, int64(0))

	})
}
