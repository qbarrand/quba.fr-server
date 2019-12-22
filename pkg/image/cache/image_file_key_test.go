package cache

import "testing"

func TestNewImageFileKey(t *testing.T) {
	if NewImageFileKey("/test/file", 1, 2, 3, "custom_fmt") == nil {
		t.Fail()
	}
}

func TestImageFileKey_Key(t *testing.T) {
	cases := []struct {
		expected string
		got      string
	}{
		{
			expected: "/test/file_1_2_3_some_fmt",
			got:      NewImageFileKey("/test/file", 1, 2, 3, "some_fmt").FsPath(),
		},
		{
			expected: "/test/file.jpg_1_2_3_some_fmt",
			got:      NewImageFileKey("/test/file.jpg", 1, 2, 3, "some_fmt").FsPath(),
		},
	}

	for _, c := range cases {
		t.Run(c.expected, func(t *testing.T) {
			if c.expected != c.got {
				t.Fatalf("Expected %q, got %q", c.expected, c.got)
			}
		})
	}
}
