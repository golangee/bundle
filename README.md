# bundle
is a workaround until [35950](https://github.com/golang/go/issues/35950) is resolved. 
*bundle* embedds files, using a few wanted optimizations, and there is no other 
which supports this out of the box:
* give direct access to (cached) uncompressed, brotli and gzip streams
* optionally cache variants in memory (e.g. for web servers)
* uses one of the most efficient codecs (brotli), which compresses 14-21% better than gzip at
comparable decompression speed
* uses base85 instead base64 to reduce embedding and parsing overhead from 33% to 25%. It is unclear
if using [base-122](http://blog.kevinalbs.com/base122) would be a good choice. The resulting go-file
has only around 21% overhead, if compressing again with bzip.
* optimized http handler which uses etags and no-cache headers 
and optimized in-memory caches of compression variants
* only regenerates source code, if files have changed. Perfect for *go generate*.


## alternatives
there are so many...

* fileembed - https://godoc.org/perkeep.org/pkg/fileembed 
* packr - https://godoc.org/github.com/gobuffalo/packr
* stuffbin - https://godoc.org/github.com/knadh/stuffbin
* vfsgen - https://github.com/shurcooL/vfsgen
* go.rice - https://github.com/GeertJohan/go.rice
* statik - https://github.com/rakyll/statik
* esc - https://github.com/mjibson/esc
* go-embed - https://github.com/pyros2097/go-embed
* go-resources - https://github.com/omeid/go-resources
* statics - https://github.com/go-playground/statics
* templify - https://github.com/wlbr/templify
* gnoso/go-bindata - https://github.com/gnoso/go-bindata
* shuLhan/go-bindata - https://github.com/shuLhan/go-bindata
* fileb0x - https://github.com/UnnoTed/fileb0x
* gobundle - https://github.com/alecthomas/gobundle
* parcello - https://github.com/phogolabs/parcello