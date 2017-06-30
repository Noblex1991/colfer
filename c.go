package colfer

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pascaldekloe/name"
)

const cKeywords = "auto break case char const continue default do double else enum extern float for goto if int long register return short signed sizeof static struct switch typedef union unsigned void volatile while"

// IsCKeyword returs whether s is a reserved word in C code.
func IsCKeyword(s string) bool {
	for _, k := range strings.Fields(cKeywords) {
		if k == s {
			return true
		}
	}
	return false
}

// GenerateC writes the code into file "Colfer.h" and "Colfer.c".
func GenerateC(basedir string, packages Packages) error {
	for _, p := range packages {
		for _, s := range p.Structs {
			s.NameNative = name.SnakeCase(p.Name + "_" + s.Name)

			for _, f := range s.Fields {
				f.NameNative = name.SnakeCase(f.Name)
				if IsCKeyword(f.NameNative) {
					f.NameNative += "_"
				}

				switch f.Type {
				case "bool":
					f.TypeNative = "char"
				case "uint8", "uint16", "uint32", "uint64", "int32", "int64":
					f.TypeNative = f.Type + "_t"
				case "float32":
					f.TypeNative = "float"
				case "float64":
					f.TypeNative = "double"
				case "timestamp":
					f.TypeNative = "timespec"
				case "binary", "text":
					f.TypeNative = "colfer_" + f.Type
				}
			}
		}
	}

	if err := os.MkdirAll(basedir, os.ModeDir|os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(basedir, "Colfer.h"))
	if err != nil {
		return err
	}
	if err := template.Must(template.New("C-header").Parse(cHeaderTemplate)).Execute(f, packages); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}

	f, err = os.Create(filepath.Join(basedir, "Colfer.c"))
	if err != nil {
		return err
	}
	if err := template.Must(template.New("C").Parse(cTemplate)).Execute(f, packages); err != nil {
		return err
	}
	return f.Close()
}

const cHeaderTemplate = `// Code generated by colf(1); DO NOT EDIT.
{{- range .}}
// The compiler used schema file {{.SchemaFileList}} for package {{.Name}}.
{{- end}}

#ifndef COLFER_H
#define COLFER_H

#include <limits.h>
#include <stdint.h>
#include <string.h>
{{- if .HasTimestamp}}
#include <time.h>
{{end}}

#if CHAR_BIT != 8
#error "octet byte size"
#endif

#ifdef __cplusplus
extern "C" {
#endif


// colfer_size_max is the upper limit for serial octet sizes.
extern size_t colfer_size_max;

// colfer_list_max is the upper limit for the number of elements in a list.
extern size_t colfer_list_max;


// colfer_text is a UTF-8 CLOB.
typedef struct {
	const char*  utf8;
	size_t       len;
} colfer_text;

// colfer_binary is a BLOB.
typedef struct {
	uint8_t* octets;
	size_t   len;
} colfer_binary;

{{range .}}{{range .Structs}}
typedef struct {{.NameNative}} {{.NameNative}};
{{end}}{{end}}
{{range .}}{{range .Structs}}
{{.DocText "// "}}
struct {{.NameNative}} {
{{- range .Fields}}
{{.DocText "\t// "}}{{- if .TypeList}}
 {{- if eq .Type "float32"}}
	struct {
		float* list;
		size_t len;
	}
 {{- else if eq .Type "float64"}}
	struct {
		double* list;
		size_t len;
	}
 {{- else if eq .Type "binary" "text"}}
	struct {
		colfer_{{.Type}}* list;
		size_t len;
	}
 {{- else}}
	struct {
		struct {{.TypeRef.NameNative}}* list;
		size_t len;
	}
 {{- end}}
{{- else}}
 {{- if eq .Type "timestamp"}}
	struct {{.TypeNative}}
 {{- else if .TypeRef}}
	{{.TypeRef.NameNative}}*
 {{- else}}
	{{.TypeNative}}
 {{- end}}
{{- end}} {{.NameNative}};
{{- end}}
};

// {{.NameNative}}_marshal_len returns the Colfer serial octet size.
// When the return is zero then errno is set to EFBIG to indicate a breach of
// either colfer_size_max or colfer_list_max.
size_t {{.NameNative}}_marshal_len(const {{.NameNative}}* o);

// {{.NameNative}}_marshal encodes o as Colfer into buf and returns the number
// of octets written.
size_t {{.NameNative}}_marshal(const {{.NameNative}}* o, void* buf);

// {{.NameNative}}_unmarshal decodes data as Colfer into o and returns the
// number of octets read. The data is read up to a maximum of datalen or
// colfer_size_max, whichever occurs first.
// When the return is zero then errno is set to one of the following 3 values:
// EWOULDBLOCK on incomplete data, EFBIG on a breach of either colfer_size_max
// or colfer_list_max and EILSEQ on schema mismatch.
size_t {{.NameNative}}_unmarshal({{.NameNative}}* o, const void* data, size_t datalen);
{{end}}{{end}}

#ifdef __cplusplus
} // extern "C"
#endif

#endif
`

const cTemplate = `// Code generated by colf(1); DO NOT EDIT.
{{- range .}}
// The compiler used schema file {{.SchemaFileList}} for package {{.Name}}.
{{- end}}

#include "Colfer.h"
#include <errno.h>
#include <stdlib.h>
{{- if .HasTimestamp}}
#include <time.h>
{{end}}

#if defined(__BYTE_ORDER) && __BYTE_ORDER == __BIG_ENDIAN || \
    defined(__BIG_ENDIAN__) || \
    defined(__ARMEB__) || \
    defined(__AARCH64EB__) || \
    defined(_MIPSEB) || defined(__MIPSEB) || defined(__MIPSEB__) || \
    defined(__SYSC_ZARCH__)
#define COLFER_ENDIAN
#endif

{{with index . 0}}
size_t colfer_size_max = {{.SizeMax}};
size_t colfer_list_max = {{.ListMax}};
{{end}}

{{range .}}{{range .Structs}}
size_t {{.NameNative}}_marshal_len(const {{.NameNative}}* o) {
	size_t l = 1;
{{range .Fields}}{{if eq .Type "bool"}}
	if (o->{{.NameNative}}) l++;
{{else if eq .Type "uint8"}}
	if (o->{{.NameNative}}) l += 2;
{{else if eq .Type "uint16"}}
	{
		uint_fast16_t x = o->{{.NameNative}};
		if (x) l += x < 256 ? 2 : 3;
	}
{{else if eq .Type "uint32"}}
	{
		uint_fast32_t x = o->{{.NameNative}};
		if (x) {
			if (x >= (uint_fast32_t) 1 << 21) l += 5;
			else for (l += 2; x > 127; x >>= 7, ++l);
		}
	}
{{else if eq .Type "uint64"}}
	{
		uint_fast64_t x = o->{{.NameNative}};
		if (x) {
			if (x >= (uint_fast64_t) 1 << 49) l += 9;
			else for (l += 2; x > 127; x >>= 7, ++l);
		}
	}
{{else if eq .Type "int32"}}
	{
		uint_fast32_t x = o->{{.NameNative}};
		if (x) {
			if (x & (uint_fast32_t) 1 << 31) {
				x = ~x;
				++x;
			}
			for (l += 2; x > 127; x >>= 7, ++l);
		}
	}
{{else if eq .Type "int64"}}
	{
		uint_fast64_t x = o->{{.NameNative}};
		if (x) {
			if (x & (uint_fast64_t) 1 << 63) {
				x = ~x;
				++x;
			}
			size_t max = l + 10;
			for (l += 2; x > 127 && l < max; x >>= 7, ++l);
		}
	}
{{else if eq .Type "float32"}}
 {{- if not .TypeList}}
	if (o->{{.NameNative}} != 0.0f) l += 5;
 {{- else}}
	{
		size_t n = o->{{.NameNative}}.len;
		if (n) {
			if (n > colfer_list_max) {
				errno = EFBIG;
				return 0;
			}
			for (l += n * 4 + 2; n > 127; n >>= 7, ++l);
		}
	}
 {{- end}}
{{else if eq .Type "float64"}}
 {{- if not .TypeList}}
	if (o->{{.NameNative}} != 0.0) l += 9;
 {{- else}}
	{
		size_t n = o->{{.NameNative}}.len;
		if (n) {
			if (n > colfer_list_max) {
				errno = EFBIG;
				return 0;
			}
			for (l += n * 8 + 2; n > 127; n >>= 7, ++l);
		}
	}
 {{- end}}
{{else if eq .Type "timestamp"}}
	{
		time_t s = o->{{.NameNative}}.tv_sec;
		long ns = o->{{.NameNative}}.tv_nsec;
		if (s || ns) {
			s += ns / 1000000000;
			l += s >= (time_t) 1 << 32 || s < 0 ? 13 : 9;
		}
	}
{{else if eq .Type "text"}}
 {{- if not .TypeList}}
	{
		size_t n = o->{{.NameNative}}.len;
		if (n > colfer_size_max) {
			errno = EFBIG;
			return 0;
		}
		if (n) for (l += 2 + n; n > 127; n >>= 7, ++l);
	}
 {{- else}}
	{
		size_t n = o->{{.NameNative}}.len;
		if (n) {
			if (n > colfer_list_max) {
				errno = EFBIG;
				return 0;
			}
			colfer_text* a = o->{{.NameNative}}.list;
			for (size_t i = 0; i < n; ++i) {
				size_t len = a[i].len;
				if (len > colfer_size_max) {
					errno = EFBIG;
					return 0;
				}
				for (l += len + 1; len > 127; len >>= 7, ++l);
			}
			for (l += 2; n > 127; n >>= 7, ++l);
			if (l > colfer_size_max) {
				errno = EFBIG;
				return 0;
			}
		}
	}
 {{- end}}
{{else if eq .Type "binary"}}
 {{- if not .TypeList}}
	{
		size_t n = o->{{.NameNative}}.len;
		if (n > colfer_size_max) {
			errno = EFBIG;
			return 0;
		}
		if (n) for (l += 2 + n; n > 127; n >>= 7, ++l);
	}
 {{- else}}
	{
		size_t n = o->{{.NameNative}}.len;
		if (n) {
			if (n > colfer_list_max) {
				errno = EFBIG;
				return 0;
			}
			colfer_binary* a = o->{{.NameNative}}.list;
			for (size_t i = 0; i < n; ++i) {
				size_t len = a[i].len;
				if (len > colfer_size_max) {
					errno = EFBIG;
					return 0;
				}
				for (l += len + 1; len > 127; len >>= 7, ++l);
			}
			for (l += 2; n > 127; n >>= 7, ++l);
			if (l > colfer_size_max) {
				errno = EFBIG;
				return 0;
			}
		}
	}
 {{- end}}
{{else}}
 {{- if not .TypeList}}
	{
		if (o->{{.NameNative}}) l += 1 + {{.TypeRef.NameNative}}_marshal_len(o->{{.NameNative}});
	}
 {{- else}}
	{
		size_t n = o->{{.NameNative}}.len;
		if (n) {
			if (n > colfer_list_max) {
				errno = EFBIG;
				return 0;
			}
			{{.TypeRef.NameNative}}* a = o->{{.NameNative}}.list;
			for (size_t i = 0; i < n; ++i) l += {{.TypeRef.NameNative}}_marshal_len(&a[i]);
			for (l += 2; n > 127; n >>= 7, ++l);
			if (l > colfer_size_max) {
				errno = EFBIG;
				return 0;
			}
		}
	}
 {{- end}}
{{end}}{{end}}
	if (l > colfer_size_max) {
		errno = EFBIG;
		return 0;
	}
	return l;
}

size_t {{.NameNative}}_marshal(const {{.NameNative}}* o, void* buf) {
	// octet pointer navigation
	uint8_t* p = buf;
{{range .Fields}}{{if eq .Type "bool"}}
	if (o->{{.NameNative}}) *p++ = {{.Index}};
{{else if eq .Type "uint8"}}
	if (o->{{.NameNative}}) {
		*p++ = {{.Index}};

		*p++ = o->{{.NameNative}};
	}
{{else if eq .Type "uint16"}}
	{
		uint_fast16_t x = o->{{.NameNative}};
		if (x) {
			if (x < 256)  {
				*p++ = {{.Index}} | 0x80;

				*p++ = x;
			} else {
				*p++ = {{.Index}};

				*p++ = x >> 8;
				*p++ = x;
			}
		}
	}
{{else if eq .Type "uint32"}}
	{
		uint_fast32_t x = o->{{.NameNative}};
		if (x) {
			if (x < (uint_fast32_t) 1 << 21) {
				*p++ = {{.Index}};
				for (; x >= 128; x >>= 7) *p++ = x | 128;
				*p++ = x;
			} else {
				*p++ = {{.Index}} | 128;
#ifdef COLFER_ENDIAN
				memcpy(p, &o->{{.NameNative}}, 4);
				p += 4;
#else
				*p++ = x >> 24;
				*p++ = x >> 16;
				*p++ = x >> 8;
				*p++ = x;
#endif
			}
		}
	}
{{else if eq .Type "uint64"}}
	{
		uint_fast64_t x = o->{{.NameNative}};
		if (x) {
			if (x < (uint_fast64_t) 1 << 49) {
				*p++ = {{.Index}};
				for (; x >= 128; x >>= 7) *p++ = x | 128;
				*p++ = x;
			} else {
				*p++ = {{.Index}} | 128;
#ifdef COLFER_ENDIAN
				memcpy(p, &o->{{.NameNative}}, 8);
				p += 8;
#else
				*p++ = x >> 56;
				*p++ = x >> 48;
				*p++ = x >> 40;
				*p++ = x >> 32;
				*p++ = x >> 24;
				*p++ = x >> 16;
				*p++ = x >> 8;
				*p++ = x;
#endif
			}
		}
	}
{{else if eq .Type "int32"}}
	{
		uint_fast32_t x = o->{{.NameNative}};
		if (x) {
			if (x & (uint_fast32_t) 1 << 31) {
				*p++ = {{.Index}} | 128;
				x = ~x + 1;
			} else	*p++ = {{.Index}};

			for (; x >= 128; x >>= 7) *p++ = x | 128;
			*p++ = x;
		}
	}
{{else if eq .Type "int64"}}
	{
		uint_fast64_t x = o->{{.NameNative}};
		if (x) {
			if (x & (uint_fast64_t) 1 << 63) {
				*p++ = {{.Index}} | 128;
				x = ~x + 1;
			} else	*p++ = {{.Index}};

			uint8_t* max = p + 8;
			for (; x >= 128 && p < max; x >>= 7) *p++ = x | 128;
			*p++ = x;
		}
	}
{{else if eq .Type "float32"}}
 {{- if not .TypeList}}
	if (o->{{.NameNative}} != 0.0f) {
		*p++ = {{.Index}};

#ifdef COLFER_ENDIAN
		memcpy(p, &o->{{.NameNative}}, 4);
		p += 4;
#else
		uint_fast32_t x;
		memcpy(&x, &o->{{.NameNative}}, 4);
		*p++ = x >> 24;
		*p++ = x >> 16;
		*p++ = x >> 8;
		*p++ = x;
#endif
	}
 {{- else}}
	{
		size_t n = o->{{.NameNative}}.len;
		if (n) {
			*p++ = {{.Index}};

			uint_fast32_t x = n;
			for (; x >= 128; x >>= 7) *p++ = x | 128;
			*p++ = x;

#ifdef COLFER_ENDIAN
			memcpy(p, o->{{.NameNative}}.list, n * 4);
			p += n * 4;
#else
			uint32_t* fp = (uint32_t*) o->{{.NameNative}}.list;
			for (;;) {
				memcpy(&x, fp, 4);
				*p++ = x >> 24;
				*p++ = x >> 16;
				*p++ = x >> 8;
				*p++ = x;
				if (--n == 0) break;
				++fp;
			}
#endif
		}
	}
 {{- end}}
{{else if eq .Type "float64"}}
 {{- if not .TypeList}}
	if (o->{{.NameNative}} != 0.0) {
		*p++ = {{.Index}};

#ifdef COLFER_ENDIAN
		memcpy(p, &o->{{.NameNative}}, 8);
		p += 8;
#else
		uint_fast64_t x;
		memcpy(&x, &o->{{.NameNative}}, 8);
		*p++ = x >> 56;
		*p++ = x >> 48;
		*p++ = x >> 40;
		*p++ = x >> 32;
		*p++ = x >> 24;
		*p++ = x >> 16;
		*p++ = x >> 8;
		*p++ = x;
#endif
	}
 {{- else}}
	{
		size_t n = o->{{.NameNative}}.len;
		if (n) {
			*p++ = {{.Index}};

			uint_fast32_t x = n;
			for (; x >= 128; x >>= 7) *p++ = x | 128;
			*p++ = x;

#ifdef COLFER_ENDIAN
			memcpy(p, o->{{.NameNative}}.list, n * 8);
			p += n * 8;
#else
			uint64_t* fp = (uint64_t*) o->{{.NameNative}}.list;
			for (;;) {
				uint_fast64_t x;
				memcpy(&x, fp, 8);
				*p++ = x >> 56;
				*p++ = x >> 48;
				*p++ = x >> 40;
				*p++ = x >> 32;
				*p++ = x >> 24;
				*p++ = x >> 16;
				*p++ = x >> 8;
				*p++ = x;
				if (--n == 0) break;
				++fp;
			}
#endif
		}
	}
 {{- end}}
{{else if eq .Type "timestamp"}}
	{
		time_t s = o->{{.NameNative}}.tv_sec;
		long ns = o->{{.NameNative}}.tv_nsec;
		if (s || ns) {
			static const int_fast64_t nano = 1000000000;
			s += ns / nano;
			ns %= nano;
			if (ns < 0) {
				--s;
				ns += nano;
			}

			uint_fast64_t x = s;
			if (x < (uint_fast64_t) 1 << 32)
				*p++ = {{.Index}};
			else {
				*p++ = {{.Index}} | 128;

				*p++ = x >> 56;
				*p++ = x >> 48;
				*p++ = x >> 40;
				*p++ = x >> 32;
			}
			*p++ = x >> 24;
			*p++ = x >> 16;
			*p++ = x >> 8;
			*p++ = x;

			x = ns;
			*p++ = x >> 24;
			*p++ = x >> 16;
			*p++ = x >> 8;
			*p++ = x;
		}
	}
{{else if eq .Type "text"}}
 {{- if not .TypeList}}
	{
		size_t n = o->{{.NameNative}}.len;
		if (n) {
			*p++ = {{.Index}};

			uint_fast32_t x = n;
			for (; x >= 128; x >>= 7) *p++ = x | 128;
			*p++ = x;

			memcpy(p, o->{{.NameNative}}.utf8, n);
			p += n;
		}
	}
 {{- else}}
	{
		size_t count = o->{{.NameNative}}.len;
		if (count) {
			*p++ = {{.Index}};

			uint_fast32_t x = count;
			for (; x >= 128; x >>= 7) *p++ = x | 128;
			*p++ = x;

			colfer_text* text = o->{{.NameNative}}.list;
			do {
				size_t n = text->len;
				for (x = n; x >= 128; x >>= 7) *p++ = x | 128;
				*p++ = x;

				memcpy(p, text->utf8, n);
				p += n;

				++text;
			} while (--count != 0);
		}
	}
 {{- end}}
{{else if eq .Type "binary"}}
 {{- if not .TypeList}}
	{
		size_t n = o->{{.NameNative}}.len;
		if (n) {
			*p++ = {{.Index}};

			uint_fast32_t x = n;
			for (; x >= 128; x >>= 7) *p++ = x | 128;
			*p++ = x;

			memcpy(p, o->{{.NameNative}}.octets, n);
			p += n;
		}
	}
 {{- else}}
	{
		size_t count = o->{{.NameNative}}.len;
		if (count) {
			*p++ = {{.Index}};

			uint_fast32_t x = count;
			for (; x >= 128; x >>= 7) *p++ = x | 128;
			*p++ = x;

			colfer_binary* binary = o->{{.NameNative}}.list;
			do {
				size_t n = binary->len;
				for (x = n; x >= 128; x >>= 7) *p++ = x | 128;
				*p++ = x;

				memcpy(p, binary->octets, n);
				p += n;

				++binary;
			} while (--count != 0);
		}
	}
 {{- end}}
{{else}}
 {{- if not .TypeList}}
	{
		if (o->{{.NameNative}}) {
			*p++ = {{.Index}};

			p += {{.TypeRef.NameNative}}_marshal(o->{{.NameNative}}, p);
		}
	}
 {{- else}}
	{
		size_t n = o->{{.NameNative}}.len;
		if (n) {
			*p++ = {{.Index}};

			uint_fast32_t x = n;
			for (; x >= 128; x >>= 7) *p++ = x | 128;
			*p++ = x;

			{{.TypeRef.NameNative}}* a = o->{{.NameNative}}.list;
			for (size_t i = 0; i < n; ++i) p += {{.TypeRef.NameNative}}_marshal(&a[i], p);
		}
	}
 {{- end}}
{{end}}{{end}}
	*p++ = 127;

	return p - (uint8_t*) buf;
}

size_t {{.NameNative}}_unmarshal({{.NameNative}}* o, const void* data, size_t datalen) {
	// octet pointer navigation
	const uint8_t* p = data;
	const uint8_t* end;
	int enderr;
	if (datalen < colfer_size_max) {
		end = p + datalen;
		enderr = EWOULDBLOCK;
	} else {
		end = p + colfer_size_max;
		enderr = EFBIG;
	}

	if (p >= end) {
		errno = enderr;
		return 0;
	}
	uint_fast8_t header = *p++;
{{range .Fields}}{{if eq .Type "bool"}}
	if (header == {{.Index}}) {
		o->{{.NameNative}} = 1;
		if (p >= end) {
			errno = enderr;
			return 0;
		}
		header = *p++;
	}
{{else if eq .Type "uint8"}}
	if (header == {{.Index}}) {
		if (p+1 >= end) {
			errno = enderr;
			return 0;
		}
		o->{{.NameNative}} = *p++;
		header = *p++;
	}
{{else if eq .Type "uint16"}}
	if (header == {{.Index}}) {
		if (p+2 >= end) {
			errno = enderr;
			return 0;
		}
		uint_fast16_t x = *p++;
		x <<= 8;
		o->{{.NameNative}} = x | *p++;
		header = *p++;
	} else if (header == ({{.Index}} | 128)) {
		if (p+1 >= end) {
			errno = enderr;
			return 0;
		}
		o->{{.NameNative}} = *p++;
		header = *p++;
	}
{{else if eq .Type "uint32"}}
	if (header == {{.Index}}) {
		if (p+1 >= end) {
			errno = enderr;
			return 0;
		}
		uint_fast32_t x = *p++;
		if (x > 127) {
			x &= 127;
			for (int shift = 7; ; shift += 7) {
				uint_fast32_t b = *p++;
				if (p >= end) {
					errno = enderr;
					return 0;
				}
				if (b <= 127) {
					x |= b << shift;
					break;
				}
				x |= (b & 127) << shift;
			}
		}
		o->{{.NameNative}} = x;
		header = *p++;
	} else if (header == ({{.Index}} | 128)) {
		if (p+4 >= end) {
			errno = enderr;
			return 0;
		}
		uint_fast32_t x = *p++;
		x <<= 24;
		x |= (uint_fast32_t) *p++ << 16;
		x |= (uint_fast32_t) *p++ << 8;
		x |= (uint_fast32_t) *p++;
		o->{{.NameNative}} = x;
		header = *p++;
	}
{{else if eq .Type "uint64"}}
	if (header == {{.Index}}) {
		if (p+1 >= end) {
			errno = enderr;
			return 0;
		}
		uint_fast64_t x = *p++;
		if (x > 127) {
			x &= 127;
			for (int shift = 7; ; shift += 7) {
				uint_fast64_t b = *p++;
				if (p >= end) {
					errno = enderr;
					return 0;
				}
				if (b <= 127) {
					x |= b << shift;
					break;
				}
				x |= (b & 127) << shift;
			}
		}
		o->{{.NameNative}} = x;
		header = *p++;
	} else if (header == ({{.Index}} | 128)) {
		if (p+8 >= end) {
			errno = enderr;
			return 0;
		}
		uint_fast64_t x = *p++;
		x <<= 56;
		x |= (uint_fast64_t) *p++ << 48;
		x |= (uint_fast64_t) *p++ << 40;
		x |= (uint_fast64_t) *p++ << 32;
		x |= (uint_fast64_t) *p++ << 24;
		x |= (uint_fast64_t) *p++ << 16;
		x |= (uint_fast64_t) *p++ << 8;
		x |= (uint_fast64_t) *p++;
		o->{{.NameNative}} = x;
		header = *p++;
	}
{{else if eq .Type "int32"}}
	if ((header & 127) == {{.Index}}) {
		if (p+1 >= end) {
			errno = enderr;
			return 0;
		}
		uint_fast32_t x = *p++;
		if (x > 127) {
			x &= 127;
			for (int shift = 7; shift < 35; shift += 7) {
				uint_fast32_t b = *p++;
				if (p >= end) {
					errno = enderr;
					return 0;
				}
				if (b <= 127) {
					x |= b << shift;
					break;
				}
				x |= (b & 127) << shift;
			}
		}
		if (header & 128) x = ~x + 1;
		o->{{.NameNative}} = x;
		header = *p++;
	}
{{else if eq .Type "int64"}}
	if ((header & 127) == {{.Index}}) {
		if (p+1 >= end) {
			errno = enderr;
			return 0;
		}
		uint_fast64_t x = *p++;
		if (x > 127) {
			x &= 127;
			for (int shift = 7; ; shift += 7) {
				uint_fast64_t b = *p++;
				if (p >= end) {
					errno = enderr;
					return 0;
				}
				if (b <= 127 || shift == 56) {
					x |= b << shift;
					break;
				}
				x |= (b & 127) << shift;
			}
		}
		if (header & 128) x = ~x + 1;
		o->{{.NameNative}} = x;
		header = *p++;
	}
{{else if eq .Type "float32"}}
 {{- if not .TypeList}}
	if (header == {{.Index}}) {
		if (p+4 >= end) {
			errno = enderr;
			return 0;
		}
#ifdef COLFER_ENDIAN
		memcpy(&o->{{.NameNative}}, p, 4);
		p += 4;
#else
		uint_fast32_t x = *p++;
		x <<= 24;
		x |= (uint_fast32_t) *p++ << 16;
		x |= (uint_fast32_t) *p++ << 8;
		x |= (uint_fast32_t) *p++;
		memcpy(&o->{{.NameNative}}, &x, 4);
#endif
		header = *p++;
	}
 {{- else}}
	if (header == {{.Index}}) {
		if (p >= end) {
			errno = enderr;
			return 0;
		}
		size_t n = *p++;
		if (n > 127) {
			n &= 127;
			for (int shift = 7; ; shift += 7) {
				if (p >= end) {
					errno = enderr;
					return 0;
				}
				size_t c = *p++;
				if (c <= 127) {
					n |= c << shift;
					break;
				}
				n |= (c & 127) << shift;
			}
		}
		if (n > colfer_list_max) {
			errno = EFBIG;
			return 0;
		}
		if (p+n*4 >= end) {
			errno = enderr;
			return 0;
		}
		o->{{.NameNative}}.len = n;

		float* fp = malloc(n * 4);
		o->{{.NameNative}}.list = fp;
#ifdef COLFER_ENDIAN
		memcpy(fp, p, n * 4);
		p += n * 4;
#else
		for (; n; --n, ++fp) {
			uint_fast32_t x = *p++;
			x <<= 24;
			x |= (uint_fast32_t) *p++ << 16;
			x |= (uint_fast32_t) *p++ << 8;
			x |= (uint_fast32_t) *p++;
			memcpy(fp, &x, 4);
		}
#endif
		header = *p++;
	}
 {{- end}}
{{else if eq .Type "float64"}}
 {{- if not .TypeList}}
	if (header == {{.Index}}) {
		if (p+8 >= end) {
			errno = enderr;
			return 0;
		}
#ifdef COLFER_ENDIAN
		memcpy(&o->{{.NameNative}}, p, 8);
		p += 8;
#else
		uint_fast64_t x = *p++;
		x <<= 56;
		x |= (uint_fast64_t) *p++ << 48;
		x |= (uint_fast64_t) *p++ << 40;
		x |= (uint_fast64_t) *p++ << 32;
		x |= (uint_fast64_t) *p++ << 24;
		x |= (uint_fast64_t) *p++ << 16;
		x |= (uint_fast64_t) *p++ << 8;
		x |= (uint_fast64_t) *p++;
		memcpy(&o->{{.NameNative}}, &x, 8);
#endif
		header = *p++;
	}
 {{- else}}
	if (header == {{.Index}}) {
		if (p >= end) {
			errno = enderr;
			return 0;
		}
		size_t n = *p++;
		if (n > 127) {
			n &= 127;
			for (int shift = 7; ; shift += 7) {
				if (p >= end) {
					errno = enderr;
					return 0;
				}
				size_t c = *p++;
				if (c <= 127) {
					n |= c << shift;
					break;
				}
				n |= (c & 127) << shift;
			}
		}
		if (n > colfer_list_max) {
			errno = EFBIG;
			return 0;
		}
		if (p+n*8 >= end) {
			errno = enderr;
			return 0;
		}
		o->{{.NameNative}}.len = n;

		double* fp = malloc(n * 8);
		o->{{.NameNative}}.list = fp;
#ifdef COLFER_ENDIAN
		memcpy(fp, p, n * 8);
		p += n * 8;
#else
		for (; n; --n, ++fp) {
			uint_fast64_t x = *p++;
			x <<= 56;
			x |= (uint_fast64_t) *p++ << 48;
			x |= (uint_fast64_t) *p++ << 40;
			x |= (uint_fast64_t) *p++ << 32;
			x |= (uint_fast64_t) *p++ << 24;
			x |= (uint_fast64_t) *p++ << 16;
			x |= (uint_fast64_t) *p++ << 8;
			x |= (uint_fast64_t) *p++;
			memcpy(fp, &x, 8);
		}
#endif
		header = *p++;
	}
 {{- end}}
{{else if eq .Type "timestamp"}}
	if ((header & 127) == {{.Index}}) {
		if (header & 128) {
			if (p+12 >= end) {
				errno = enderr;
				return 0;
			}
			uint64_t x = *p++;
			x <<= 56;
			x |= (uint64_t) *p++ << 48;
			x |= (uint64_t) *p++ << 40;
			x |= (uint64_t) *p++ << 32;
			x |= (uint64_t) *p++ << 24;
			x |= (uint64_t) *p++ << 16;
			x |= (uint64_t) *p++ << 8;
			x |= (uint64_t) *p++;
			o->{{.NameNative}}.tv_sec = (time_t)(int64_t) x;
		} else {
			if (p+8 >= end) {
				errno = enderr;
				return 0;
			}
			uint_fast32_t x = *p++;
			x <<= 24;
			x |= (uint_fast32_t) *p++ << 16;
			x |= (uint_fast32_t) *p++ << 8;
			x |= (uint_fast32_t) *p++;
			o->{{.NameNative}}.tv_sec = (time_t) x;
		}
		uint_fast32_t x = *p++;
		x <<= 24;
		x |= (uint_fast32_t) *p++ << 16;
		x |= (uint_fast32_t) *p++ << 8;
		x |= (uint_fast32_t) *p++;
		o->{{.NameNative}}.tv_nsec = (long) x;
		header = *p++;
	}
{{else if eq .Type "text"}}
 {{- if not .TypeList}}
	if (header == {{.Index}}) {
		if (p >= end) {
			errno = enderr;
			return 0;
		}
		size_t n = *p++;
		if (n > 127) {
			n &= 127;
			for (int shift = 7; shift < sizeof(size_t) * CHAR_BIT; shift += 7) {
				if (p >= end) {
					errno = enderr;
					return 0;
				}
				size_t c = *p++;
				if (c <= 127) {
					n |= c << shift;
					break;
				}
				n |= (c & 127) << shift;
			}
		}
		if (n > colfer_size_max) {
			errno = EFBIG;
			return 0;
		}
		if (p+n >= end) {
			errno = enderr;
			return 0;
		}
		o->{{.NameNative}}.len = n;

		void* a = malloc(n);
		o->{{.NameNative}}.utf8 = (char*) a;
		if (n) {
			memcpy(a, p, n);
			p += n;
		}
		header = *p++;
	}
 {{- else}}
	if (header == {{.Index}}) {
		if (p >= end) {
			errno = enderr;
			return 0;
		}
		size_t n = *p++;
		if (n > 127) {
			n &= 127;
			for (int shift = 7; ; shift += 7) {
				if (p >= end) {
					errno = enderr;
					return 0;
				}
				size_t c = *p++;
				if (c <= 127) {
					n |= c << shift;
					break;
				}
				n |= (c & 127) << shift;
			}
		}
		if (n > colfer_list_max) {
			errno = EFBIG;
			return 0;
		}
		o->{{.NameNative}}.len = n;

		colfer_text* text = malloc(n * sizeof(colfer_text));
		o->{{.NameNative}}.list = text;
		for (; n; --n, ++text) {
			if (p >= end) {
				errno = enderr;
				return 0;
			}
			size_t len = *p++;
			if (len > 127) {
				len &= 127;
				for (int shift = 7; ; shift += 7) {
					if (p >= end) {
						errno = enderr;
						return 0;
					}
					size_t c = *p++;
					if (c <= 127) {
						len |= c << shift;
						break;
					}
					len |= (c & 127) << shift;
				}
			}
			if (len > colfer_size_max) {
				errno = EFBIG;
				return 0;
			}
			if (p+len >= end) {
				errno = enderr;
				return 0;
			}
			text->len = len;

			char* a = malloc(len);
			text->utf8 = a;
			if (len) {
				memcpy(a, p, len);
				p += len;
			}
		}

		if (p >= end) {
			errno = enderr;
			return 0;
		}
		header = *p++;
	}
 {{- end}}
{{else if eq .Type "binary"}}
 {{- if not .TypeList}}
	if (header == {{.Index}}) {
		if (p >= end) {
			errno = enderr;
			return 0;
		}
		size_t n = *p++;
		if (n > 127) {
			n &= 127;
			for (int shift = 7; ; shift += 7) {
				if (p >= end) {
					errno = enderr;
					return 0;
				}
				size_t c = *p++;
				if (c <= 127) {
					n |= c << shift;
					break;
				}
				n |= (c & 127) << shift;
			}
		}
		if (n > colfer_size_max) {
			errno = EFBIG;
			return 0;
		}
		if (p+n >= end) {
			errno = enderr;
			return 0;
		}
		o->{{.NameNative}}.len = n;

		void* a = malloc(n);
		o->{{.NameNative}}.octets = (uint8_t*) a;
		if (n) {
			memcpy(a, p, n);
			p += n;
		}
		header = *p++;
	}
 {{- else}}
	if (header == {{.Index}}) {
		if (p >= end) {
			errno = enderr;
			return 0;
		}
		size_t n = *p++;
		if (n > 127) {
			n &= 127;
			for (int shift = 7; ; shift += 7) {
				if (p >= end) {
					errno = enderr;
					return 0;
				}
				size_t c = *p++;
				if (c <= 127) {
					n |= c << shift;
					break;
				}
				n |= (c & 127) << shift;
			}
		}
		if (n > colfer_list_max) {
			errno = EFBIG;
			return 0;
		}
		o->{{.NameNative}}.len = n;

		colfer_binary* binary = malloc(n * sizeof(colfer_binary));
		o->{{.NameNative}}.list = binary;
		for (; n; --n, ++binary) {
			if (p >= end) {
				errno = enderr;
				return 0;
			}
			size_t len = *p++;
			if (len > 127) {
				len &= 127;
				for (int shift = 7; ; shift += 7) {
					if (p >= end) {
						errno = enderr;
						return 0;
					}
					size_t c = *p++;
					if (c <= 127) {
						len |= c << shift;
						break;
					}
					len |= (c & 127) << shift;
				}
			}
			if (len > colfer_size_max) {
				errno = EFBIG;
				return 0;
			}
			if (p+len >= end) {
				errno = enderr;
				return 0;
			}
			binary->len = len;

			uint8_t* a = malloc(len);
			binary->octets = a;
			if (len) {
				memcpy(a, p, len);
				p += len;
			}
		}

		if (p >= end) {
			errno = enderr;
			return 0;
		}
		header = *p++;
	}
 {{- end}}
{{else}}
 {{- if not .TypeList}}
	if (header == {{.Index}}) {
		o->{{.NameNative}} = calloc(1, sizeof({{.TypeRef.NameNative}}));
		size_t read = {{.TypeRef.NameNative}}_unmarshal(o->{{.NameNative}}, p, (size_t) (end - p));
		if (!read) {
			if (errno == EWOULDBLOCK) errno = enderr;
			return read;
		}
		p += read;

		if (p >= end) {
			errno = enderr;
			return 0;
		}
		header = *p++;
	}
 {{- else}}
	if (header == {{.Index}}) {
		if (p >= end) {
			errno = enderr;
			return 0;
		}
		size_t n = *p++;
		if (n > 127) {
			n &= 127;
			for (int shift = 7; ; shift += 7) {
				if (p >= end) {
					errno = enderr;
					return 0;
				}
				size_t c = *p++;
				if (c <= 127) {
					n |= c << shift;
					break;
				}
				n |= (c & 127) << shift;
			}
		}
		if (n > colfer_list_max) {
			errno = EFBIG;
			return 0;
		}

		{{.TypeRef.NameNative}}* a = calloc(n, sizeof({{.TypeRef.NameNative}}));
		for (size_t i = 0; i < n; ++i) {
			size_t read = {{.TypeRef.NameNative}}_unmarshal(&a[i], p, (size_t) (end - p));
			if (!read) {
				if (errno == EWOULDBLOCK) errno = enderr;
				return read;
			}
			p += read;
		}
		o->{{.NameNative}}.len = n;
		o->{{.NameNative}}.list = a;

		if (p >= end) {
			errno = enderr;
			return 0;
		}
		header = *p++;
	}
 {{- end}}
{{end}}{{end}}
	if (header != 127) {
		errno = EILSEQ;
		return 0;
	}

	return (size_t) (p - (const uint8_t*) data);
}
{{end}}{{end}}`
