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
			status varchar(15) default 'NEW',
			accrual double precision default 0,
			id_user int references users(id),
			uploaded_at timestamp default now()
		)
	`

	createTableUserBalances = `
		create table if not exists user_balances (
			id serial primary key not null,
			id_user int references users(id),
			balance double precision default 0,
			uploaded_at timestamp default now()
		)
	`

	createTableOperations = `
		create table if not exists operations (
			id serial primary key not null,
			balance_id int references user_balances(id),
			order_id int references orders(id),
			type varchar(10) default 'debit',
			amount double precision,
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

	updateOrder = `
		update orders set 
			status = $1,
			accrual = $2
		where number = $3
	`

	selectOrders = `
		select number, status, accrual, uploaded_at 
		from orders
		left join users on orders.id_user = users.id
		where users.user_id = $1
		order by uploaded_at desc
	`

	createDefaultUserBalance = `
		insert into user_balances (id_user) values ((select id from users where user_id = $1))
	`

	selectUserOnlyBalance = `
		select balance from user_balances 
		where id_user = (select id from users where user_id = $1)
	`

	selectUserBalance = `
		select 
			ub.balance,
			coalesce(sum(amount), 0)
		from users u
		left join user_balances ub on u.id = ub.id_user
		left join operations op on op.balance_id = ub.id
		where u.user_id = $1
		group by ub.balance
	`

	selectUserBalanceAndOrder = `
			select
			ub.balance,
			orders.number
		from users
		left join user_balances ub on ub.id_user = users.id
		left join orders on orders.id_user = users.id
		where users.user_id = $1
		and orders.number = $2
	`

	updateUserBalance = `
		update user_balances 
			set balance=$1 
		where 
			id_user = (select id from users where user_id = $2)
	`

	insertOperation = `
		insert into operations (
			balance_id, 
			order_id, 
			amount
		)
		values (
			(select id from user_balances where id_user = (select id from users where user_id = $1)),
			(select id from orders where number = $2),
			$3
		)
	`

	selectUserOperations = `
		select 
			orders.number,
			operations.amount,
			operations.uploaded_at
		from operations
		left join orders on orders.id = operations.order_id
		left join user_balances ub on ub.id = operations.balance_id
		left join users on users.id = ub.id_user
		where users.user_id = $1
		order by operations.uploaded_at desc
	`
)
