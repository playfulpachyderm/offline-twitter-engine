PRAGMA foreign_keys = on;

create table spans (rowid integer primary key,
	name text not null,
	parent_id integer references spans(rowid),
	start_time integer not null,
	end_time integer not null
);


create view fq_spans as
	with recursive fq_name(rowid, root_id, fq_name, name, parent_id, start_time, end_time, duration) as (
	  select rowid, rowid as root_id, name, name, parent_id, start_time, end_time, end_time - start_time
	    from spans
	   where parent_id is null

	  union all

	  select s.rowid, fq.root_id, fq.fq_name || '/' || s.name, s.name, s.parent_id, s.start_time, s.end_time, s.end_time - s.start_time
	    from spans s
	    join fq_name fq on s.parent_id = fq.rowid
	)

	select rowid, root_id, fq_name, name, parent_id, start_time, end_time, duration
	  from fq_name;


-- Meta
-- ----

create table db_version(rowid integer primary key,
    version integer not null unique
);
insert into db_version(version) values (0);
