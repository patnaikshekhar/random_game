CREATE TABLE USERS (
  Id bigserial primary key, 
  email text, 
  password text
);

INSERT INTO USERS (email, password) 
VALUES ('1@a.com', '1');

INSERT INTO USERS (email, password) 
VALUES ('2@a.com', '2');

CREATE TABLE GAMES (
  Id bigserial primary key, 
  player1 bigint references USERS, 
  player2 bigint references USERS, 
  status text DEFAULT 'Started'
);

CREATE TABLE SHIPS (
  Id bigserial primary key, 
  gameID bigint references GAMES, 
  playerID bigint references USERS, 
  size int,
  xlocation1 smallint,
  ylocation1 smallint,
  xlocation2 smallint,
  ylocation2 smallint,
  xlocation3 smallint,
  ylocation3 smallint,
  xlocation4 smallint,
  ylocation4 smallint,
  xlocation5 smallint,
  ylocation5 smallint,
  sunk boolean
);