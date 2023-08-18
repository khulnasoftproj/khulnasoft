package khulnasoft_test

import (
	"testing"

	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config/khulnasoft"
)

func TestFilterPackageByTag(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name         string
		pkg          *khulnasoft.Package
		tags         map[string]struct{}
		excludedTags map[string]struct{}
		exp          bool
	}{
		{
			name: "no tag",
			pkg: &khulnasoft.Package{
				Name:     "cli/cli",
				Version:  "v2.0.0",
				Registry: "standard",
			},
			exp: true,
		},
		{
			name: "package has tags but no tag is specified",
			pkg: &khulnasoft.Package{
				Name:     "cli/cli",
				Version:  "v2.0.0",
				Registry: "standard",
				Tags:     []string{"ci"},
			},
			exp: true,
		},
		{
			name: "tag is matched",
			pkg: &khulnasoft.Package{
				Name:     "cli/cli",
				Version:  "v2.0.0",
				Registry: "standard",
				Tags:     []string{"ci", "foo"},
			},
			tags: map[string]struct{}{
				"ci": {},
			},
			exp: true,
		},
		{
			name: "exclude tag is matched",
			pkg: &khulnasoft.Package{
				Name:     "cli/cli",
				Version:  "v2.0.0",
				Registry: "standard",
				Tags:     []string{"ci", "foo"},
			},
			excludedTags: map[string]struct{}{
				"ci": {},
			},
			exp: false,
		},
		{
			name: "exclude tag and tag are matched",
			pkg: &khulnasoft.Package{
				Name:     "cli/cli",
				Version:  "v2.0.0",
				Registry: "standard",
				Tags:     []string{"ci", "foo"},
			},
			tags: map[string]struct{}{
				"foo": {},
			},
			excludedTags: map[string]struct{}{
				"ci": {},
			},
			exp: false,
		},
		{
			name: "exclude tag isn't matched and tag is matched",
			pkg: &khulnasoft.Package{
				Name:     "cli/cli",
				Version:  "v2.0.0",
				Registry: "standard",
				Tags:     []string{"ci", "foo"},
			},
			tags: map[string]struct{}{
				"foo": {},
			},
			excludedTags: map[string]struct{}{
				"yoo": {},
			},
			exp: true,
		},
		{
			name: "exclude tag and tag aren't matched",
			pkg: &khulnasoft.Package{
				Name:     "cli/cli",
				Version:  "v2.0.0",
				Registry: "standard",
				Tags:     []string{"ci"},
			},
			tags: map[string]struct{}{
				"foo": {},
			},
			excludedTags: map[string]struct{}{
				"yoo": {},
			},
			exp: false,
		},
		{
			name: "exclude tag ins't matched and tag isn't specified",
			pkg: &khulnasoft.Package{
				Name:     "cli/cli",
				Version:  "v2.0.0",
				Registry: "standard",
				Tags:     []string{"ci"},
			},
			excludedTags: map[string]struct{}{
				"yoo": {},
			},
			exp: true,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			f := khulnasoft.FilterPackageByTag(d.pkg, d.tags, d.excludedTags)
			if f != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, f)
			}
		})
	}
}
