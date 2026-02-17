package cli

// Migration running has been moved to the core package.
// Users should now run migrations from their main.go file by importing migration packages
// and calling core.RunMigrations(app).
//
// Example:
//   import (
//       "github.com/ishubhamsingh2e/bourbon/bourbon/core"
//       _ "yourproject/apps/users/migrations"
//   )
//
//   func main() {
//       app := core.NewApplication("./settings.toml")
//       app.ConnectDB()
//
//       // Run migrations
//       if err := core.RunMigrations(app); err != nil {
//           log.Fatal(err)
//       }
//
//       app.Run()
//   }

