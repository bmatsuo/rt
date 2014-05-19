package rt

import "testing"

func TestCompose(t *testing.T) {
	for i, test := range []struct{ pat, param, path string }{
		{"/", "hello", "/hello"},
		{"/hello", "world", "/hello"},
		{"/hello/", "world", "/hello/world"},
		{"/hello/", "", "/hello/"},
	} {
		path := Compose(test.pat, test.param)
		if path != test.path {
			t.Errorf("test %d: Compose(%q, %q) == %q (!= %q)", i, test.pat, test.param, path, test.path)
		}
	}
}

func TestDecompose(t *testing.T) {
	for i, test := range []struct{ pat, path, param string }{
		{"/", "/hello", "hello"},
		{"/hello/", "/hello/world", "world"},
		{"/hello", "/hello/world", ""},
	} {
		param := Decompose(test.pat, test.path)
		if param != test.param {
			t.Errorf("test %d: Deompose(%q, %q) == %q (!= %q)", i, test.pat, test.path, param, test.param)
		}
	}
}
