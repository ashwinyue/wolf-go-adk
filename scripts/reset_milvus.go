//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"log"

	milvusClient "github.com/milvus-io/milvus-sdk-go/v2/client"
)

func main() {
	ctx := context.Background()

	// 连接 Milvus
	cli, err := milvusClient.NewClient(ctx, milvusClient.Config{
		Address: "localhost:19530",
	})
	if err != nil {
		log.Fatalf("连接 Milvus 失败: %v", err)
	}
	defer cli.Close()

	// 删除 collection
	exists, err := cli.HasCollection(ctx, "wolf_episodes")
	if err != nil {
		log.Fatalf("检查 Collection 失败: %v", err)
	}

	if exists {
		fmt.Println("🗑️ 删除旧的 Collection 'wolf_episodes'...")
		if err := cli.DropCollection(ctx, "wolf_episodes"); err != nil {
			log.Fatalf("删除 Collection 失败: %v", err)
		}
		fmt.Println("✅ 删除成功")
	} else {
		fmt.Println("ℹ️ Collection 'wolf_episodes' 不存在")
	}

	fmt.Println("\n下次运行游戏时会自动创建新的 Collection")
}
