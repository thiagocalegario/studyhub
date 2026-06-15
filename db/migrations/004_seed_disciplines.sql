INSERT INTO disciplines (code, name, semester, workload) VALUES
-- 1º Semestre
('CIC0004', 'Algoritmos e Programação de Computadores', 1, 90),
('CIC0005', 'Formação Docente em Computação', 1, 60),
('PAD0028', 'Organização da Educação Brasileira', 1, 60),

-- 2º Semestre
('CIC0002', 'Fundamentos Teóricos da Computação', 2, 60),
('CIC0090', 'Estruturas de Dados', 2, 60),
('MAT0025', 'Cálculo 1', 2, 90),
('TEF0011', 'Psicologia da Educação', 2, 60),

-- 3º Semestre
('CIC0182', 'Lógica Computacional 1', 3, 60),
('CIC0197', 'Técnicas de Programação 1', 3, 60),
('EST0022', 'Probabilidade e Estatística', 3, 90),
('MAT0031', 'Introdução à Álgebra Linear', 3, 60),

-- 4º Semestre
('CIC0092', 'Organização de Arquivos', 4, 60),
('CIC0093', 'Linguagens de Programação', 4, 60),
('CIC0177', 'Arquitetura de Processadores Digitais', 4, 60),
('CIC0206', 'Métodos de Pesquisa na Licenciatura em Computação', 4, 30),
('MAT0026', 'Cálculo 2', 4, 90),

-- 5º Semestre
('CIC0097', 'Bancos de Dados', 5, 60),
('CIC0207', 'Projeto Interdisciplinar de Licenciatura em Computação', 5, 60),
('CIC0208', 'Produção de Material Didático', 5, 75),
('CIC0209', 'Supervisão de Produção de Material Didático', 5, 30),
('MTC0012', 'Didática Fundamental', 5, 60),

-- 6º Semestre
('CIC0189', 'Projeto e Análise de Algoritmos', 6, 60),
('CIC0210', 'Prática Pedagógica em Computação 1', 6, 90),
('CIC0212', 'Supervisão de Prática Pedagógica em Computação 1', 6, 30),

-- 7º Semestre
('CIC0101', 'Sistemas de Informação', 7, 60),
('CIC0105', 'Engenharia de Software', 7, 60),
('CIC0201', 'Segurança Computacional', 7, 60),
('CIC0211', 'Prática Pedagógica em Computação 2', 7, 90),
('CIC0213', 'Supervisão de Prática Pedagógica em Computação 2', 7, 30),
('CIC0224', 'Projeto de Pesquisa na Licenciatura em Computação', 7, 30),

-- 8º Semestre
('CIC0158', 'Informática Aplicada à Educação', 8, 60),
('CIC0214', 'Estágio Supervisionado em Licenciatura em Computação 1', 8, 105),
('CIC0217', 'Supervisão de Estágio em Licenciatura em Computação 1', 8, 30),
('CIC0223', 'Produção Científica na Licenciatura em Computação', 8, 30),
('LIP0174', 'Língua de Sinais Brasileira - Básico', 8, 60),

-- 9º Semestre
('CIC0215', 'Estágio Supervisionado em Licenciatura em Computação 2', 9, 105),
('CIC0216', 'Estágio Supervisionado em Licenciatura em Computação 3', 9, 105),
('CIC0219', 'Supervisão de Estágio em Licenciatura em Computação 2', 9, 30),
('CIC0221', 'Supervisão de Estágio em Licenciatura em Computação 3', 9, 30)

ON CONFLICT DO NOTHING;