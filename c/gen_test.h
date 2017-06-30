#include "gen/Colfer.h"

#include <math.h>
#include <stdint.h>


typedef struct golden {
	const char* hex;
	const gen_o o;
} golden;

const struct golden golden_cases[] = {
	{"7f", {.b = 0}},
	{"007f", {.b = 1}},
	{"01017f", {.u32 = 1}},
	{"01ff017f", {.u32 = UINT8_MAX}},
	{"01ffff037f", {.u32 = UINT16_MAX}},
	{"81ffffffff7f", {.u32 = UINT32_MAX}},
	{"02017f", {.u64 = 1}},
	{"02ff017f", {.u64 = UINT8_MAX}},
	{"02ffff037f", {.u64 = UINT16_MAX}},
	{"02ffffffff0f7f", {.u64 = UINT32_MAX}},
	{"82ffffffffffffffff7f", {.u64 = UINT64_MAX}},
	{"03017f", {.i32 = 1}},
	{"83017f", {.i32 = -1}},
	{"037f7f", {.i32 = INT8_MAX}},
	{"8380017f", {.i32 = INT8_MIN}},
	{"03ffff017f", {.i32 = INT16_MAX}},
	{"838080027f", {.i32 = INT16_MIN}},
	{"03ffffffff077f", {.i32 = INT32_MAX}},
	{"8380808080087f", {.i32 = INT32_MIN}},
	{"04017f", {.i64 = 1}},
	{"84017f", {.i64 = -1}},
	{"047f7f", {.i64 = INT8_MAX}},
	{"8480017f", {.i64 = INT8_MIN}},
	{"04ffff017f", {.i64 = INT16_MAX}},
	{"848080027f", {.i64 = INT16_MIN}},
	{"04ffffffff077f", {.i64 = INT32_MAX}},
	{"8480808080087f", {.i64 = INT32_MIN}},
	{"04ffffffffffffffff7f7f", {.i64 = INT64_MAX}},
	{"848080808080808080807f", {.i64 = INT64_MIN}},
	{"05000000017f", {.f32 = 1e-45}},
	{"057f7fffff7f", {.f32 = 3.4028235e+38}},
	{"057fc000007f", {.f32 = NAN}},
	{"0600000000000000017f", {.f64 = 5e-324}},
	{"067fefffffffffffff7f", {.f64 = 1.7976931348623157e+308}},
	{"067ff80000000000007f", {.f64 = NAN}},
	{"0755ef312a2e5da4e77f", {.t = {.tv_sec = 1441739050, .tv_nsec = 777888999}}},
	{"87000007dba8218000000003e87f", {.t = {.tv_sec = 864E10, .tv_nsec = 1000}}},
	{"87fffff82457de8000000003e97f", {.t = {.tv_sec = -864E10, .tv_nsec = 1001}}},
	{"87ffffffffffffffff2e5da4e77f", {.t = {.tv_sec = -1, .tv_nsec = 777888999}}},
	{"0801417f", {.s = {.utf8 = "A", .len = 1}}},
	{"080261007f", {.s = {.utf8 = "a\x00", .len = 2}}},
	{"0809c280e0a080f09080807f", {.s = {.utf8 = "\xc2\x80\xe0\xa0\x80\xf0\x90\x80\x80", .len = 9}}},
	{"08800120202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020207f",
		{.s = {.utf8 = "                                                                                                                                ", .len = 128}}},
	{"0901ff7f", {.a = {.octets = (uint8_t*) "\xff", .len = 1}}},
	{"090202007f", {.a = {.octets = (uint8_t*) "\x02\x00", .len = 2}}},
	{"09c0010909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909090909097f",
		{.a = {.octets = (uint8_t*) "\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09\x09", .len = 192}}},
	{"0a7f7f", {.o = &((gen_o) {.b = 0})}},
	{"0a007f7f", {.o = &((gen_o) {.b = 1})}},
	{"0b01007f7f", {.os = {.list = (gen_o[1]) {((gen_o) {.b = 1})}, .len = 1}}},
	{"0b027f7f7f", {.os = {.list = (gen_o[2]) {((gen_o) {.b = 0}), ((gen_o) {.b = 0})}, .len = 2}}},
	{"0c0300016101627f", {.ss = {.list = (colfer_text[3]) {{.utf8 = "", .len = 0}, {.utf8 = "a", .len = 1}, {.utf8 = "b", .len = 1}}, .len = 3 }}},
	{"0d0201000201027f", {.as = {.list = (colfer_binary[2]) {{.octets = (uint8_t*) "\x00", .len = 1}, {.octets = (uint8_t*) "\x01\x02", .len = 2}}, .len = 2}}},
	{"0e017f", {.u8 = 1}},
	{"0eff7f", {.u8 = UINT8_MAX}},
	{"8f017f", {.u16 = 1}},
	{"0fffff7f", {.u16 = UINT16_MAX}},
	{"1002000000003f8000007f", {.f32s = {.list = (float[2]) {0.0f, 1.0f}, .len = 2}}},
	{"11014058c000000000007f", {.f64s = {.list = (double[1]) {99.0}, .len = 1}}}
};
