# 并发库

## （一） 消费者和生产者
    cg :=NewConcurrencyGet(10,2)
    	go func() {
    		for i:=0; i < 100;i++ {
    			cg.Push(i)
    		}
    		cg.CloseProduct()
    	}()
        
        // index 表示第几个协程，用来连接对象
    	cg.Pop(func(index int, v interface{}) error {
        	_ = dbpool[index] //数据连接数组
            log.Info(index, v)
    		return nil
    	})
    	cg.Wait()