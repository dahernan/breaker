package breaker

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSuccessCounter(t *testing.T) {
	Convey("Success keeps the counter right", t, func() {
		hc, err := NewHealthCounts(3)
		So(err, ShouldBeNil)

		sum := hc.Summary()
		So(sum.Success, ShouldEqual, int64(0))
		So(sum.Failures, ShouldEqual, int64(0))
		So(sum.Total, ShouldEqual, int64(0))
		So(sum.ErrorPercentage, ShouldEqual, 0.0)

		hc.Success()
		hc.Success()
		hc.Success()

		sum = hc.Summary()
		So(sum.Success, ShouldEqual, int64(3))
		So(sum.Failures, ShouldEqual, int64(0))
		So(sum.Total, ShouldEqual, int64(3))
		So(sum.ErrorPercentage, ShouldEqual, 0.0)

		hc.Success()

		sum = hc.Summary()
		So(sum.Success, ShouldEqual, int64(4))
		So(sum.Failures, ShouldEqual, int64(0))
		So(sum.Total, ShouldEqual, int64(4))
		So(sum.ErrorPercentage, ShouldEqual, 0.0)
	})

}

func TestFailuresCounter(t *testing.T) {
	Convey("Failures keeps the counter right", t, func() {
		hc, err := NewHealthCounts(3)
		So(err, ShouldBeNil)

		hc.Fail()
		hc.Fail()

		sum := hc.Summary()
		So(sum.Success, ShouldEqual, int64(0))
		So(sum.Failures, ShouldEqual, int64(2))
		So(sum.Total, ShouldEqual, int64(2))
		So(sum.ErrorPercentage, ShouldEqual, 100.0)

		hc.Fail()
		hc.Fail()
		hc.Fail()

		sum = hc.Summary()
		So(sum.Success, ShouldEqual, int64(0))
		So(sum.Failures, ShouldEqual, int64(5))
		So(sum.Total, ShouldEqual, int64(5))
		So(sum.ErrorPercentage, ShouldEqual, 100.0)
	})
}

func TestErrorPercentage(t *testing.T) {
	Convey("The Error Percentage is calculated", t, func() {
		hc, err := NewHealthCounts(3)
		So(err, ShouldBeNil)

		hc.Success()
		hc.Success()
		hc.Fail()
		hc.Fail()

		sum := hc.Summary()
		So(sum.Success, ShouldEqual, int64(2))
		So(sum.Failures, ShouldEqual, int64(2))
		So(sum.Total, ShouldEqual, int64(4))
		So(sum.ErrorPercentage, ShouldEqual, 50.0)

		hc.Fail()

		sum = hc.Summary()
		So(sum.Success, ShouldEqual, int64(2))
		So(sum.Failures, ShouldEqual, int64(3))
		So(sum.Total, ShouldEqual, int64(5))
		So(sum.ErrorPercentage, ShouldEqual, 60.0)

		hc.Success()
		hc.Success()

		sum = hc.Summary()
		So(sum.Success, ShouldEqual, int64(4))
		So(sum.Failures, ShouldEqual, int64(3))
		So(sum.Total, ShouldEqual, int64(7))
		So(sum.ErrorPercentage, ShouldEqual, (3.0/7.0)*100)
	})
}

func TestRotateBucketsForgetsOldResults(t *testing.T) {
	Convey("Internal storage rotates the buckets to write", t, func() {
		hc, err := NewHealthCounts(3)
		So(err, ShouldBeNil)

		hc.Success()
		hc.Success()
		hc.Fail()
		hc.Fail()

		sum := hc.Summary()
		So(sum.Success, ShouldEqual, int64(2))
		So(sum.Failures, ShouldEqual, int64(2))
		So(sum.Total, ShouldEqual, int64(4))
		So(sum.ErrorPercentage, ShouldEqual, 50.0)

		// force to write in the second bucket
		time.Sleep(1 * time.Second)

		hc.Success()
		sum = hc.Summary()
		So(sum.Success, ShouldEqual, int64(3))
		So(sum.Failures, ShouldEqual, int64(2))
		So(sum.Total, ShouldEqual, int64(5))
		So(sum.ErrorPercentage, ShouldEqual, 40.0)

		// force to write in the third bucket
		time.Sleep(1 * time.Second)

		hc.Success()
		hc.Success()
		hc.Fail()
		hc.Fail()

		sum = hc.Summary()
		So(sum.Success, ShouldEqual, int64(5))
		So(sum.Failures, ShouldEqual, int64(4))
		So(sum.Total, ShouldEqual, int64(9))
		So(sum.ErrorPercentage, ShouldEqual, (4.0/9.0)*100)

		// force to forget the result of the first bucket
		time.Sleep(1 * time.Second)

		sum = hc.Summary()
		So(sum.Success, ShouldEqual, int64(3))
		So(sum.Failures, ShouldEqual, int64(2))
		So(sum.Total, ShouldEqual, int64(5))
		So(sum.ErrorPercentage, ShouldEqual, (2.0/5.0)*100)
	})
}

func TestRotateBucketsResetOldBuckets(t *testing.T) {
	Convey("Internal storage resets old buckets", t, func() {
		hc, err := NewHealthCounts(3)
		So(err, ShouldBeNil)

		hc.Success()
		hc.Success()
		hc.Fail()
		hc.Fail()

		sum := hc.Summary()
		So(sum.Success, ShouldEqual, int64(2))
		So(sum.Failures, ShouldEqual, int64(2))
		So(sum.Total, ShouldEqual, int64(4))
		So(sum.ErrorPercentage, ShouldEqual, 50.0)

		// force to write in the second bucket
		time.Sleep(1 * time.Second)

		hc.Success()
		sum = hc.Summary()
		So(sum.Success, ShouldEqual, int64(3))
		So(sum.Failures, ShouldEqual, int64(2))
		So(sum.Total, ShouldEqual, int64(5))
		So(sum.ErrorPercentage, ShouldEqual, 40.0)

		// force to write in the third bucket
		time.Sleep(1 * time.Second)

		hc.Success()
		hc.Success()
		hc.Fail()
		hc.Fail()

		sum = hc.Summary()
		So(sum.Success, ShouldEqual, int64(5))
		So(sum.Failures, ShouldEqual, int64(4))
		So(sum.Total, ShouldEqual, int64(9))
		So(sum.ErrorPercentage, ShouldEqual, (4.0/9.0)*100)

		// force to overwrite the first bucket
		time.Sleep(1 * time.Second)

		hc.Success()

		sum = hc.Summary()
		So(sum.Success, ShouldEqual, int64(4))
		So(sum.Failures, ShouldEqual, int64(2))
		So(sum.Total, ShouldEqual, int64(6))
		So(sum.ErrorPercentage, ShouldEqual, (float64(2.0)/float64(6.0))*100.0)
	})
}

func TestHealthCountsIsAbleToShutdown(t *testing.T) {
	Convey("Stops works properly", t, func() {
		hc, err := NewHealthCounts(3)
		So(err, ShouldBeNil)

		hc.Success()
		hc.Success()

		// now shutdown
		hc.Cancel()

		_, ok := <-hc.ctx.Done()
		So(ok, ShouldEqual, false)
	})
}

func TestNewHealthCountsShouldGetAnErrorWithMoreThan60Seconds(t *testing.T) {
	Convey("Returns an error if you try to storage more than 60 seconds", t, func() {
		_, err := NewHealthCounts(61)
		So(err, ShouldEqual, ErrNumberOfSecondsToStoreOutOfBounds)

		_, err = NewHealthCounts(60)
		So(err, ShouldBeNil)
	})
}

func TestNewHealthCountsShouldGetAnErrorWith0Seconds(t *testing.T) {
	Convey("Returns an error if you try to storage 0 or negative seconds", t, func() {
		_, err := NewHealthCounts(0)
		So(err, ShouldEqual, ErrNumberOfSecondsToStoreOutOfBounds)

		_, err = NewHealthCounts(-1)
		So(err, ShouldEqual, ErrNumberOfSecondsToStoreOutOfBounds)

		_, err = NewHealthCounts(1)
		So(err, ShouldBeNil)
	})
}
