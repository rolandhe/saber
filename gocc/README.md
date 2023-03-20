提供类似java juc的库能力。包括：

* Cond with timeout，相比于标准库Cond，它最典型的特点是支持timeout
* Future，类似java的Future，提交到任务执行器的任务可以被异步执行，调用者持有Future来获取异步执行的结果或者取消任务
* FutureGroup, 包含多个Future，可以在FutureGroup上等待所有的Future任务都执行完成，也可以取消任务，相比于在多个Future上一个个轮询，调用更加简单
* BlockingQueue, 支持并发调用的、并行安全的队列，强制有界
* Executor, 用于异步执行任务的执行器，强制指定并发数。要执行的任务提交给Executor后马上返回Future，调用者持有Future来获取最终结果，Executor内执行完成任务或者发现任务取消后会修改Future的内部状态
* Semaphore，信号量
* CountdownLatch， 倒计数