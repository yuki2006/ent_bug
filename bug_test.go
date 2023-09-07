package bug

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"testing"

	"entgo.io/bug/ent/user"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"entgo.io/bug/ent"
	"entgo.io/bug/ent/enttest"
)

func TestBugSQLite(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()
	test(t, client)
}

func TestBugMySQL(t *testing.T) {
	for version, port := range map[string]int{"56": 3306, "57": 3307, "8": 3308} {
		addr := net.JoinHostPort("localhost", strconv.Itoa(port))
		t.Run(version, func(t *testing.T) {
			client := enttest.Open(t, dialect.MySQL, fmt.Sprintf("root:pass@tcp(%s)/test?parseTime=True", addr))
			defer client.Close()
			test(t, client)
		})
	}
}

func TestBugPostgres(t *testing.T) {
	for version, port := range map[string]int{"10": 5430, "11": 5431, "12": 5432, "13": 5433, "14": 5434} {
		t.Run(version, func(t *testing.T) {
			client := enttest.Open(t, dialect.Postgres, fmt.Sprintf("host=localhost port=%d user=postgres dbname=test password=pass sslmode=disable", port))
			defer client.Close()
			test(t, client)
		})
	}
}

func TestBugMaria(t *testing.T) {
	for version, port := range map[string]int{"10.5": 4306, "10.2": 4307, "10.3": 4308} {
		t.Run(version, func(t *testing.T) {
			addr := net.JoinHostPort("localhost", strconv.Itoa(port))
			client := enttest.Open(t, dialect.MySQL, fmt.Sprintf("root:pass@tcp(%s)/test?parseTime=True", addr))
			defer client.Close()
			test(t, client)
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

	users := []*ent.User{
		{Name: "Alice", Age: nil},
		{Name: "Bob", Age: Ptr(25)},
	}

	for _, u := range users {
		_, err = tx.User.Create().SetName(u.Name).SetNillableAge(u.Age).Save(ctx)
		if err != nil {
			log.Fatalf("failed creating user: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("failed committing transaction: %v", err)
	}

	tx, err = client.Tx(ctx)
	if err != nil {
		panic(err)
		return
	}
	all, err := tx.User.Query().Order(user.ByAge(sql.OrderNullsFirst())).All(ctx)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	fmt.Printf("all: %v\n", all)
}

func Ptr[T any](i T) *T {
	return &i
}
