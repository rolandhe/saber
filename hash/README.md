# golang cityhash实现，关于cityhash算法参照 https://github.com/google/cityhash
本实现完全移植c版本。

* 支持32位
* 支持64位
* 支持128位
* 支持256位

同样使用sse4.2 simd函数，对应_mm_crc32_u64函数的调用使用汇编实现。参照crc/*的代码。

# 示例：

## 32 位的hash

```
	hash := CityHash32String("我们将通过生成一个大的文件的方式来检验各种方法的执行效率因为这种方式在结束的时候需要执行文件")
```

## 64位hash

```
    hash :=  := CityHash32String("我们将通过生成一个大的文件的方式来检验各种方法的执行效率因为这种方式在结束的时候需要执行文件")
```