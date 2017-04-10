
# Simplefs

按照 Facebook Haystack 论文用 Golang 实现的文件存储系统, 仅作学习用.

> 论文原文: [facebook haystack](http://www.usenix.org/event/osdi10/tech/full_papers/Beaver.pdf)

## 组件

* api. 用来测试和展示效果的 HTTP API
* core. 对 Haystack 中 `needle`, `directory`, `store` 的实现
* main. `Main` 函数
* utils. 其它功能

## 参考

* [whalefs](https://github.com/030io/whalefs)
* [seaweedfs](https://github.com/chrislusf/seaweedfs)


## License

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.