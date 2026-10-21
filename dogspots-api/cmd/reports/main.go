// Comando de moderación: lista las denuncias registradas ordenadas de más reciente a más antigua.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/angellsan94/dogspots-api/internal/reports"
	_ "modernc.org/sqlite"
)

func main() {
	dbPath := flag.String("db", "dogspots.db", "ruta del fichero SQLite")
	limit := flag.Int("limit", 100, "número máximo de denuncias a mostrar")
	flag.Parse()

	database, err := sql.Open("sqlite", "file:"+*dbPath+"?_pragma=busy_timeout(5000)")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error al abrir la base de datos: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	items, err := reports.NewSQLiteRepository(database).List(context.Background(), *limit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error al listar denuncias: %v\n", err)
		os.Exit(1)
	}
	if len(items) == 0 {
		fmt.Println("No hay denuncias.")
		return
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "FECHA\tTIPO\tDESTINO\tMOTIVO\tDENUNCIANTE\tDETALLES\tID")
	for _, item := range items {
		kind, target := "desconocido", item.ID.String()
		switch {
		case item.PlaceID != nil:
			kind, target = "lugar", item.PlaceID.String()+" "+item.PlaceTitle
		case item.CommentID != nil:
			kind, target = "comentario", item.CommentID.String()+" "+snippet(item.CommentBody)
		}
		details := "-"
		if item.Details != nil {
			details = snippet(*item.Details)
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			item.CreatedAt.Format("2006-01-02 15:04"),
			kind,
			target,
			item.Reason,
			item.ReporterEmail,
			details,
			item.ID,
		)
	}
	if err := writer.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "error al escribir la salida: %v\n", err)
		os.Exit(1)
	}
}

func snippet(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	runes := []rune(value)
	if len(runes) > 60 {
		return string(runes[:60]) + "…"
	}
	return value
}
