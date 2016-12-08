// Code generated by protoc-gen-go.
// source: poseidon_if.proto
// DO NOT EDIT!

/*
Package datastruct is a generated protocol buffer package.

It is generated from these files:
	poseidon_if.proto

It has these top-level messages:
	DocGzMeta
	DocId
	DocIdList
	CompressedDocIdList
	InvertedIndex
	CompressedInvertedIndex
	InvertedIndexGzMeta
*/
package datastruct

import proto "github.com/golang/protobuf/proto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal

// 原始数据按照gz压缩文件格式存放在hdfs中
// 每128行原始数据合在一起称为一个 Document（文档）
// 一个hdfs文件按照2GB大小计算，大约可以容纳 10w 个压缩后的 Document
// 我们用 DocGzMeta 结构来描述文档相关的元数据信息
type DocGzMeta struct {
	Path   string `protobuf:"bytes,1,opt,name=path" json:"path,omitempty"`
	Offset uint32 `protobuf:"varint,2,opt,name=offset" json:"offset,omitempty"`
	Length uint32 `protobuf:"varint,3,opt,name=length" json:"length,omitempty"`
}

func (m *DocGzMeta) Reset()         { *m = DocGzMeta{} }
func (m *DocGzMeta) String() string { return proto.CompactTextString(m) }
func (*DocGzMeta) ProtoMessage()    {}

type DocId struct {
	DocId    uint32 `protobuf:"varint,1,opt,name=docId" json:"docId,omitempty"`
	RowIndex uint32 `protobuf:"varint,2,opt,name=rowIndex" json:"rowIndex,omitempty"`
}

func (m *DocId) Reset()         { *m = DocId{} }
func (m *DocId) String() string { return proto.CompactTextString(m) }
func (*DocId) ProtoMessage()    {}

// 一个分词可能会出现多个文档中，由于每个文档有多行原始数据组成
// 每个关联数据需要 docId、rawIndex 两个信息来描述
type DocIdList struct {
	// 该分词所关联的 Document ID。按照 docId 升序排列
	// 为了方便 protobuf 的 varint 压缩存储，采用差分数据来存储
	// 差分数据：后一个数据的存储值等于它的原始值减去前一个数据的原始
	// 举例如下：
	// 假如原始 docId 列表为：1,3,4,7,9,115,120,121,226
	// 那么实际存储的数据为： 1,2,1,3,2,106,6,1,105
	DocIds []*DocId `protobuf:"bytes,1,rep,name=docIds" json:"docIds,omitempty"`
}

func (m *DocIdList) Reset()         { *m = DocIdList{} }
func (m *DocIdList) String() string { return proto.CompactTextString(m) }
func (*DocIdList) ProtoMessage()    {}

func (m *DocIdList) GetDocIds() []*DocId {
	if m != nil {
		return m.DocIds
	}
	return nil
}

// 压缩的docIdList, 使用FastPFOR算法压缩，两个数组解压后等长
type CompressedDocIdList struct {
	DocList []uint32 `protobuf:"varint,1,rep,name=docList" json:"docList,omitempty"`
	RowList []uint32 `protobuf:"varint,2,rep,name=rowList" json:"rowList,omitempty"`
}

func (m *CompressedDocIdList) Reset()         { *m = CompressedDocIdList{} }
func (m *CompressedDocIdList) String() string { return proto.CompactTextString(m) }
func (*CompressedDocIdList) ProtoMessage()    {}

// Token->DocIds 倒排索引表结构。这个索引数据最终每天需要占用2TB
// hashid=hash64(token)%100亿，重复(冲突)不影响
// 直接在hdfs上进行分词，中间数据文件(按照hashid排序，总共100亿行)：hashid token list<DocId>
//
// 索引文件创建过程
//      loop:
//          1. 取N行(N=1000)，生成一个 InvertedIndex 对象，序列化，gz压缩，追加到hdfs文件中
//             记录: hdfspath hashid offset length
//          2. 如果 hashid%N == M (M=1000 具体取值可以参考hdfs文件大小等于256MB左右的时候，为宜），重新写一个新的hdfs文件
//             N*M*277(每个token对应的DocIdList.Item个数)*4(每个DocIdList.Item占用4自己)*0.2(压缩比) --> N=1000,M=1000结果在256M以内
//          3. 回到1
//      上述第1步中记录的4个字段，hdfspath、hashid可以根据规则推测出来，因此只需要记录offset、length即可
//      总共需要记录 1000w (=总分词数/N)，每个8字节，总计需要80M，这个文件可以存放在hdfs中，加载的时候可以加载到缓存中(redis)
type InvertedIndex struct {
	Index map[string]*DocIdList `protobuf:"bytes,1,rep,name=index" json:"index,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *InvertedIndex) Reset()         { *m = InvertedIndex{} }
func (m *InvertedIndex) String() string { return proto.CompactTextString(m) }
func (*InvertedIndex) ProtoMessage()    {}

func (m *InvertedIndex) GetIndex() map[string]*DocIdList {
	if m != nil {
		return m.Index
	}
	return nil
}

type CompressedInvertedIndex struct {
	Index map[string]*CompressedDocIdList `protobuf:"bytes,1,rep,name=index" json:"index,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *CompressedInvertedIndex) Reset()         { *m = CompressedInvertedIndex{} }
func (m *CompressedInvertedIndex) String() string { return proto.CompactTextString(m) }
func (*CompressedInvertedIndex) ProtoMessage()    {}

func (m *CompressedInvertedIndex) GetIndex() map[string]*CompressedDocIdList {
	if m != nil {
		return m.Index
	}
	return nil
}

type InvertedIndexGzMeta struct {
	Offset uint32 `protobuf:"varint,1,opt,name=offset" json:"offset,omitempty"`
	Length uint32 `protobuf:"varint,2,opt,name=length" json:"length,omitempty"`
	Path   string `protobuf:"bytes,3,opt,name=path" json:"path,omitempty"`
}

func (m *InvertedIndexGzMeta) Reset()         { *m = InvertedIndexGzMeta{} }
func (m *InvertedIndexGzMeta) String() string { return proto.CompactTextString(m) }
func (*InvertedIndexGzMeta) ProtoMessage()    {}

func init() {
}
