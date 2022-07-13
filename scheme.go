package main

import (
	"fmt"
	"io/fs"
	"strings"
)

type ColorScheme struct {
	Name        string
	System      string
	Author      string
	Slug        string
	Description string
	Palette     map[string]color
}

func (s *ColorScheme) TemplateVariables() map[string]interface{} {
	ret := map[string]interface{}{
		"scheme-name":             s.Name,
		"scheme-author":           s.Author,
		"scheme-slug":             s.Slug,
		"scheme-system":           s.System,
		"scheme-description":      s.Description,
		"scheme-slug-underscored": strings.Replace(s.Slug, "-", "_", -1),
	}

	for colorKey, colorVal := range s.Palette {
		// Note that we only lowercase the output of this to match the reference
		// repo.
		ret[colorKey+"-hex"] = fmt.Sprintf("%02x%02x%02x", colorVal.R, colorVal.G, colorVal.B)
		ret[colorKey+"-hex-bgr"] = fmt.Sprintf("%02x%02x%02x", colorVal.B, colorVal.G, colorVal.R)

		ret[colorKey+"-rgb-r"] = colorVal.R
		ret[colorKey+"-rgb-g"] = colorVal.G
		ret[colorKey+"-rgb-b"] = colorVal.B
		ret[colorKey+"-dec-r"] = float32(colorVal.R) / 255
		ret[colorKey+"-dec-g"] = float32(colorVal.G) / 255
		ret[colorKey+"-dec-b"] = float32(colorVal.B) / 255
		ret[colorKey+"-hex-r"] = fmt.Sprintf("%02x", colorVal.R)
		ret[colorKey+"-hex-g"] = fmt.Sprintf("%02x", colorVal.G)
		ret[colorKey+"-hex-b"] = fmt.Sprintf("%02x", colorVal.B)
	}

	return ret
}

func loadSchemes(schemesFS fs.FS) ([]*ColorScheme, bool) {
	schemes := make(map[string]map[string]*ColorScheme)

	// Walk the fs.FS we have and load all yaml files as scheme files.
	err := fs.WalkDir(schemesFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		filename := d.Name()
		if !strings.HasSuffix(filename, ".yaml") {
			return nil
		}

		scheme, err := LoadScheme(schemesFS, path)
		if err != nil {
			return err
		}

		if _, ok := schemes[scheme.System]; !ok {
			schemes[scheme.System] = make(map[string]*ColorScheme)
		}

		if _, ok := schemes[scheme.System][scheme.Slug]; ok {
			log.WithField("scheme", scheme.Slug).Warnf("Conflicting scheme")
		}

		log.Debugf("Found scheme %q", scheme.Slug)

		schemes[scheme.System][scheme.Slug] = scheme

		return nil
	})
	if err != nil {
		log.Error(err)
		return nil, false
	}

	// Flatten all the schemes into a list.
	var ret []*ColorScheme
	for _, system := range schemes {
		for _, scheme := range system {
			ret = append(ret, scheme)
		}
	}

	return ret, true
}
