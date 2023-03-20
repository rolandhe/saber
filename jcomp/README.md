# 与java兼容的字符串处理
## jstring
golang的string底层是utf8，而java是utf16， 二者的len不同，substring也不同。由于java的utf16可以包含大部分的中文字符，所以基于java string的length已经实现了大量业务，因此golang需要兼容java的行为，保证二者的一致性。
jstring提供如下能力：
* JavaStringLen, 返回和java String.Length兼容的长度
* JavaSubString， 实现Java String.substring(int start, int end)语义
* JavaSubStringToEnd， 实现Java String.substring(int start)语义
* JavaToChars， 实现Java Character.toChars(int codepoint)语义
* JavaCharCount，实现Java Character.charCount(int codepoint)语义
* JavaCodePoint/JavaCodePointAt/JavaCodePoint, 把[]Char转成codepoint


### CodePoint和Char
codepoint是unicode的概念，任何一个字符在unicode体系中都有对应的一个table 行的位置，这个位置用数字表示，就是codepoint，它用一个32bit int来描述，但到目前为止仅仅使用了24bit。codepoint在golang中用rune来描述。
在java中每个codepoint用一个或2个char来描述，最多两个。每个char对应0~65535, 这和golang的uint16正好相对应。

所以jstring声明了CodePoint和Char类型，他们分别是rune和uint16的别名。

### 示例

```
func javaLength(){
	s := "刘德华 andi lou"
	l, _ := jcomp.JavaStringLen(s)

	fmt.Println(l, len(s))
}
```
