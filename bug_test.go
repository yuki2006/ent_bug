package bug

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"testing"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql/schema"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"entgo.io/bug/ent"
	"entgo.io/bug/ent/enttest"
	"entgo.io/bug/ent/user"
)

func TestOrderBug(t *testing.T) {
	for version, port := range map[string]int{"56": 3306} {
		addr := net.JoinHostPort("localhost", strconv.Itoa(port))
		t.Run(version, func(t *testing.T) {
			client := enttest.Open(t, dialect.MySQL, fmt.Sprintf("root:pass@tcp(%s)/test?parseTime=True", addr)).Debug()
			defer client.Close()
			test(t, client)
			client.ExecContext(context.Background(), "DROP TABLE user")
		})
	}
}

func test(t *testing.T, client *ent.Client) {

	ctx := context.Background()
	err := client.Schema.Create(ctx, schema.WithDropIndex(true), schema.WithDropColumn(true), schema.WithForeignKeys(true))
	if err != nil {
		t.Fatal(err)
		return
	}
	client.User.Delete().ExecX(ctx)
	ctx = context.Background()
	tx, err := client.Tx(ctx)
	if err != nil {
		log.Fatalf("failed creating transaction: %v", err)
	}

	var users []*ent.User
	for i := 0; i < 30; i++ {
		users = append(users, &ent.User{
			Name: fmt.Sprintf("user%v", i), Priority: Ptr(1)},
		)
		users = append(users, &ent.User{
			Name: fmt.Sprintf("user%v", i)},
		)
	}

	for _, u := range users {
		_, err = tx.User.Create().SetName(u.Name).SetNillablePriority(u.Priority).Save(ctx)
		if err != nil {
			log.Fatalf("failed creating user: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("failed committing transaction: %v", err)
	}

	// dummy for debug
	// SELECT `users`.`id`, `users`.`priority`, `users`.`name` FROM `users` WHERE `users`.`priority` IS NULL args=[]
	client.User.Query().Where(user.PriorityIsNil()).AllX(ctx)

	limit := Ptr(5)

	part1, err := resolver(ctx, client, nil, limit, nil, nil)
	if err != nil {
		panic(err)
	}
	for _, edge := range part1.Edges {
		fmt.Printf("%+v\n", edge.Node)
	}
	fmt.Println("-------------")
	part2, err := resolver(ctx, client, nil, limit, part1.PageInfo.EndCursor, nil)
	if err != nil {
		panic(err)
	}
	if len(part2.Edges) == 0 {
		t.Fatal("part2.Edges is empty")
	}
	for _, edge := range part2.Edges {
		fmt.Printf("%+v\n", edge.Node)
	}

}

func resolver(ctx context.Context, client *ent.Client, before *ent.Cursor, first *int, after *ent.Cursor, last *int) (*ent.UserConnection, error) {
	tx, err := client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Commit()

	return tx.User.Query().Paginate(ctx, after, first, before, last,
		ent.WithUserOrder([]*ent.UserOrder{
			{
				Field:     ent.UserOrderFieldPriority,
				Direction: entgql.OrderDirectionAsc,
			},
		}),
	)
}

func Ptr[V any](v V) *V {
	return &v
}
