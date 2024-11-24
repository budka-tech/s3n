create table bucket
(
    id          smallint     not null,
    bucket_name varchar(255) not null,
    primary key (id),
    unique (bucket_name)
);

create table image
(
    id        uuid     not null,
    bucket_id smallint not null,
    primary key (id),
    foreign key (bucket_id) references bucket
        on delete restrict
);

