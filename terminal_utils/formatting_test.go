package terminal_utils_test

import (
	"testing"

	"reflect"

	"offline_twitter/terminal_utils"
)

func TestWrapParagraph(t *testing.T) {
	test_cases := []struct {
		Text     string
		Expected []string
	}{
		{
			"These are public health officials who are making decisions about your lifestyle because they know more about health, " +
				"fitness and well-being than you do",
			[]string{
				"These are public health officials who are making decisions",
				"about your lifestyle because they know more about health,",
				"fitness and well-being than you do",
			},
		},
		{
			`Things I learned in law school:`,
			[]string{`Things I learned in law school:`},
		},
		{
			`Every student is smarter than you except the ones in your group project.`,
			[]string{
				`Every student is smarter than you except the ones in your`,
				`group project.`,
			},
		},
	}
	for _, testcase := range test_cases {
		result := terminal_utils.WrapParagraph(testcase.Text, 60)
		if !reflect.DeepEqual(result, testcase.Expected) {
			t.Errorf("Expected:\n%s\nGot:\n%s\n", testcase.Expected, result)
		}
	}
}

func TestWrapText(t *testing.T) {
	test_cases := []struct {
		Text     string
		Expected string
	}{
		{
			"These are public health officials who are making decisions about your lifestyle because they know more about health, " +
				"fitness and well-being than you do",
			`These are public health officials who are making decisions
    about your lifestyle because they know more about health,
    fitness and well-being than you do`,
		},
		{
			`Things I learned in law school:
Falling behind early gives you more time to catch up.
Never use a long word when a diminutive one will suffice.
Every student is smarter than you except the ones in your group project.
If you try & fail, doesn’t matter. Try again & fail better`,
			`Things I learned in law school:
    Falling behind early gives you more time to catch up.
    Never use a long word when a diminutive one will suffice.
    Every student is smarter than you except the ones in your
    group project.
    If you try & fail, doesn’t matter. Try again & fail better`,
		},
	}
	for _, testcase := range test_cases {
		result := terminal_utils.WrapText(testcase.Text, 60)
		if result != testcase.Expected {
			t.Errorf("Expected:\n%s\nGot:\n%s\n", testcase.Expected, result)
		}
	}
}
