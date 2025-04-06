package diffs_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	diffs "gitlab.com/offline-twitter/twitter_offline_engine/pkg/webserver/tmp_templ_diff"
)

func TestEmptyTag(t *testing.T) {
	expected := "<div></div>"
	test_cases := []string{
		"<div></div>",
		"<div>\n</div>",
		"<div> </div>",
		"<div> \n  </div>",
		"<div>   </div>",
	}
	for _, test_case := range test_cases {
		diff, err := diffs.DiffStrings(expected, test_case)
		require.NoError(t, err)
		assert.Empty(t, diff, "input: %s", test_case)
	}
}

func TestJSCommentProcessing(t *testing.T) {
	expected := `
		<script>
			htmx.config.scrollBehavior = "instant";
			document.addEventListener('DOMContentLoaded', function() {
				document.body.addEventListener('htmx:beforeSwap', function(e) {
					if (e.detail.xhr.status === 500) {
						e.detail.shouldSwap = true;
						e.detail.isError = true;
					} else if (e.detail.xhr.status >= 400 && e.detail.xhr.status < 500) {
						e.detail.shouldSwap = true;
						e.detail.isError = false;
					}
				});
			});
		</script>
	`
	diff, err := diffs.DiffStrings(
		`
		<script>
// Set default scrolling ("instant", "smooth" or "auto")
htmx.config.scrollBehavior = "instant";

document.addEventListener('DOMContentLoaded', function() {
	/**
	 * Consider HTTP 4xx and 500 errors to contain valid HTMX, and swap them as usual
	 */
	document.body.addEventListener('htmx:beforeSwap', function(e) {
								if (e.detail.xhr.status === 500) {
			e.detail.shouldSwap = true;
			e.detail.isError = true;
		} else if (e.detail.xhr.status >= 400 && e.detail.xhr.status < 500) {
			e.detail.shouldSwap = true;
			e.detail.isError = false;
		}
	});
});
		</script>`,
		expected,
	)
	require.NoError(t, err)
	assert.Empty(t, diff)
}
