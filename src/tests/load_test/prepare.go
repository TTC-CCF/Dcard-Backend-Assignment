package load_test

import (
	"fmt"
	"main/models"
)

const rows = 1000

func DeleteAllData() {
	err := models.DB.Exec(`delete from banners;
	delete from countries;
	delete from genders;
	delete from platforms;`).Error

	if err != nil {
		panic(err)
	}
}

func InsertLoadTestData() {
	sql := `
	insert into genders (name) values ('M'), ('F');

	insert into countries (name) values ('TW'), ('US'), ('JP'), ('KR'), ('CN'), ('HK'), ('CA'), ('UK'), ('FR'), ('DE'), ('IT');
	insert into platforms (name) values ('ios'), ('android'), ('web');

	insert into banners (title, start_at, end_at, age_start, age_end)
	select 
		'banner ' || i,
		NOW() - INTERVAL '1 DAY' * (random() * 365)::INT,
		NOW() + INTERVAL '1 DAY' * (random() * 365)::INT,
		FLOOR((random()+i/1e39) * 50)::INT,
		FLOOR((random()+i/1e39) * 50 + 50)::INT
	from generate_series(1, %d) i;

	insert into banner_country (banner_id, country_id)
	select b.id, c.id 
	from banners b 
	join countries c on random() < 0.5;

	insert into banner_gender
	select b.id, g.id
	from banners b
	join genders g on random() < 0.5;

	insert into banner_platform
	select b.id, p.id
	from banners b
	join platforms p on random() < 0.5;`

	sql = fmt.Sprintf(sql, rows)
	if err := models.DB.Exec(sql).Error; err != nil {
		panic(err)
	}
}
