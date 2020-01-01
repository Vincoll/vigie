package utils

import "github.com/vincoll/vigie/pkg/utils/dnscache"

var ALLVARS map[string][]string
var TEMPPATH string

// TODO: Find a proper spot for this var
var CACHEDDNSRESOLVER *dnscache.Resolver
