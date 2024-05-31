INSERT INTO memecoins (id, name) VALUES (1, 'gummy');
INSERT INTO memecoins (id, name) VALUES (2, 'baked');

INSERT INTO tasks (id, name, links, descriptions, points, memecoin_id)
VALUES (1, 'follow $gummy', 'https://x.com/gummyonsol', 'this is a cool project', 200, 1);
VALUES (2, 'comment on youtube', 'https://youtube.com', 'this is a cool project', 200, 1);
VALUES (3, 'follow $baked', 'https://x.com/bakedonsol', 'this is a cool project', 200, 2);
VALUES (4, 'comment on youtube', 'https://youtube.com', 'this is a cool project', 200, 2);