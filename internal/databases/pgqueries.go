package databases

var (
	createTableUsers = `
		create table if not exists users (
			id serial primary key not null,
			login varchar(250) not null,
			password varchar(250) not null,
			user_id varchar(250)
		)
	`

	createTableOrders = `
		create table if not exists orders (
			id serial primary key not null,
			number varchar(250) not null,
			status varchar(15),
			accrual int,
			id_user int references users(id),
			uploaded_at timestamp default now()
		)
	`

	createUser = `
		insert into users (login, password, user_id)
		values ($1, $2, $3)
		returning user_id, login, password
	`

	selectUser = `
		select user_id, login, password from users where login = $1 and password = $2
	`

	checkLogin = `
		select exists (select user_id from users where login = $1) as check
	`

	checkUploadOrder = `
		select 
			users.user_id
		from orders
		left join users on orders.id_user = users.id
		where number = $1
	`

	createOrder = `
		insert into orders (number, id_user)
		values ($1, (select id from users where user_id = $2))
	`
)
