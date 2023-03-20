golang cityhash实现，关于cityhash算法参照 https://github.com/google/cityhash
本实现完全移植c版本。

* 支持32位
* 支持64位
* 支持128位
* 支持256位

同样使用sse4.2 simd函数，对应_mm_crc32_u64函数的调用使用汇编实现。参照crc/*的代码。

