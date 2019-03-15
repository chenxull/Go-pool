package Pool

//Pool 定义了一个协程池实例的所有功能
type Pool interface {
	Queue(fn WorkFunc) WorkUnit
	Reset()
	Cancel()
	Close()

	/*
		创建一个新的Batch对象，用于将工作单元与池中可能正在运行的其他工作单元分开排队。将这些工作单元分组在一起，
		可以单独取消批处理工作单元，而不会影响池中运行的任何其他工作单元，
		并且可以在完成时在通道上输出结果。
		注意:一旦QueueComplete ( )被调用，批处理就不可重复使用，
		它的生命周期被密封以完成排队的项目。
	*/
	//每个workunit都会被分配到一个Batch中
	Batch() Batch
}

//WorkFunc  需要加入pool 中执行的任务
type WorkFunc func(wu WorkUnit) (interface{}, error)
