-- name: CreatePhoto :exec
insert into photos (
    id, place_id, url, thumb_url, caption, width, height, position, uploaded_by
) values (
    ?, ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: GetPhoto :one
select id, place_id, url, thumb_url, caption, width, height, position, uploaded_by, created_at
from photos
where id = ?;

-- name: CountByPlaceID :one
select count(*) from photos where place_id = ?;

-- name: DeleteByPlaceID :exec
delete from photos where place_id = ?;

-- name: DeleteByID :exec
delete from photos where id = ?;

-- name: ListByPlaceID :many
select id, place_id, url, thumb_url, caption, width, height, position, uploaded_by, created_at
from photos
where place_id = ?
order by position asc;
