CREATE INDEX banners_time ON banners USING btree (start_at, end_at);
CREATE INDEX banners_age ON banners USING btree (age_start, age_end);
CREATE INDEX banners_country ON banners USING gin (country);
CREATE INDEX banners_gender ON banners USING gin (gender);
CREATE INDEX banners_platform ON banners USING gin (platform);
CREATE INDEX banners_country_gender ON banners USING gin (country, gender);
CREATE INDEX banners_country_platform ON banners USING gin (country, platform);
CREATE INDEX banners_gender_platform ON banners USING gin (gender, platform);
CREATE INDEX banners_country_gender_platform ON banners USING gin (country, gender, platform);