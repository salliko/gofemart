package databases

var (
	createDatabaseStruct = `
		create table if not exists users (
			id serial primary key not null,
			login varchar(250) not null,
			password varchar(250) not null,
			user_id varchar(250)
		)
	`

	createUser = `
		insert into users (login, password, user_id)
		values ($1, $2, $3)
		returning user_id, login, password
	`

	selectUser = `
		select user_id, login, password where login = $1 and password = $2
	`

	checkLogin = `
		select exists (select user_id from users where login = $1) as check
	`
)
