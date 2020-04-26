package driver

type Queue struct {
	list *SingleList
}

// Init 队列初始化
func (q *Queue) Init() {
	q.list = new(SingleList)
	q.list.Init()
}

// Enqueue 进入队列
func (q *Queue) Enqueue(data interface{}) bool {
	return q.list.Append(&SingleNode{Data: data})
}

// Dequeue 出列
func (q *Queue) Dequeue() interface{} {
	node := q.list.Get(0)
	if node == nil {
		return nil
	}
	q.list.Delete(0)
	return node.Data
}

// Peek 查看队头信息
func (q *Queue) Peek() interface{} {
	node := q.list.Get(0)
	if node == nil {
		return nil
	}
	return node.Data
}

// Size 获取队列长度
func (q *Queue) Size() uint {
	return q.list.Size
}
