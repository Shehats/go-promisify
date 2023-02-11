package promise

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testMessage struct {
	Name    string
	Subject string
}

func TestPromiseThenExecution(t *testing.T) {
	t.Run("Test then execution", func(t *testing.T) {
		p := Promisify[testMessage](func(name string, subject string) (testMessage, error) {
			return testMessage{
				Name:    name,
				Subject: subject,
			}, nil
		}, "Someone famous", "Hi famous person")
		p.Then(func(tm testMessage) {
			assert.Equal(t, tm, testMessage{
				Name:    "Someone famous",
				Subject: "Hi famous person",
			})
		})
		p.Catch(func(err error) {
			assert.Fail(t, "This should never be called")
		})
	})
	t.Run("Test then execution and creates a new promise", func(t *testing.T) {
		p := Promisify[testMessage](func(name string, subject string) (testMessage, error) {
			return testMessage{
				Name:    name,
				Subject: subject,
			}, nil
		}, "Someone famous", "Hi famous person")
		p = Then(p, func(tm testMessage) (testMessage, error) {
			assert.Equal(t, tm, testMessage{
				Name:    "Someone famous",
				Subject: "Hi famous person",
			})
			return testMessage{
				Name:    "Another famous person",
				Subject: "I am too famous to chat",
			}, nil
		})
		obj, err := p.Await()
		assert.Equal(t, obj, testMessage{
			Name:    "Another famous person",
			Subject: "I am too famous to chat",
		})
		assert.NoError(t, err)
	})
}

func TestPromiseWithCatchExecution(t *testing.T) {
	t.Run("Executes Catch on error", func(t *testing.T) {
		p := Promisify[testMessage](func(name string, subject string) (testMessage, error) {
			return testMessage{}, fmt.Errorf("Famous people don't shake hands")
		}, "Someone famous", "Hi famous person")
		p.Then(func(tm testMessage) {
			assert.Fail(t, "This block shouldn't execute")
		})
		p.Catch(func(err error) {
			// Should call this function
			assert.Error(t, err)
			assert.Equal(t, err.Error(), "Famous people don't shake hands")
		})
		p.Exec()
	})
	t.Run("Executes Catch on error and creates a new promise", func(t *testing.T) {
		p := Promisify[testMessage](func(name string, subject string) (testMessage, error) {
			return testMessage{}, fmt.Errorf("Famous people don't shake hands")
		}, "Someone famous", "Hi famous person")
		p.Then(func(tm testMessage) {
			assert.Fail(t, "This should never get called")
		})
		p1 := Catch(p, func(err error) (testMessage, error) {
			assert.Error(t, err)
			return testMessage{
				Name:    "Stunt Double",
				Subject: err.Error(),
			}, nil
		})
		msg, err := p1.Await()
		assert.NoError(t, err)
		assert.Equal(t, msg, testMessage{
			Name:    "Stunt Double",
			Subject: "Famous people don't shake hands",
		})
	})
}

func TestPromiseWithFinallyExecution(t *testing.T) {
	t.Run("Executes finally after then and catch", func(t *testing.T) {
		recovery := func() {
			if r := recover(); r != nil {
				assert.Equal(t, r, "Should panic")
			}
		}
		p := Promisify[testMessage](func(name string, subject string) (testMessage, error) {
			return testMessage{
				Name:    name,
				Subject: subject,
			}, nil
		}, "Someone famous", "Hi famous person")
		p.Then(func(tm testMessage) {
			assert.Equal(t, tm, testMessage{
				Name:    "Someone famous",
				Subject: "Hi famous person",
			})
		})
		p.Catch(func(err error) {
			assert.Fail(t, "This should never get called")
		})
		p.Finally(func() {
			defer recovery()
			panic("Should panic")
		})
	})
}
