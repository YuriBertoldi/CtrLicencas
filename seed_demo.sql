-- Equipes auxiliares
INSERT INTO auxiliares (tipo, nome) SELECT 'equipe', nome FROM (VALUES
  ('Fiscal'), ('Folha'), ('Contabil'), ('Suporte'), ('Infraestrutura')
) AS s(nome)
WHERE NOT EXISTS (SELECT 1 FROM auxiliares WHERE tipo='equipe' LIMIT 1);

-- Desenvolvedores
INSERT INTO devs (nome, equipe, status, obs) VALUES
  ('Carlos Silva',     'Fiscal',         'ativo',   'Lider tecnico do modulo fiscal'),
  ('Ana Oliveira',     'Folha',          'ativo',   'Especialista em calculos trabalhistas'),
  ('Bruno Santos',     'Contabil',       'ativo',   ''),
  ('Juliana Costa',    'Fiscal',         'ativo',   'Foco em SPED e NFe'),
  ('Rafael Mendes',    'Suporte',        'livre',   'Disponivel para novos projetos'),
  ('Fernanda Lima',    'Folha',          'ativo',   'eSocial e FGTS Digital'),
  ('Pedro Almeida',    'Infraestrutura', 'ativo',   'DevOps e automacao'),
  ('Mariana Rocha',    'Contabil',       'ativo',   'IFRS e relatorios gerenciais'),
  ('Lucas Ferreira',   'Fiscal',         'inativo', 'Desligado em maio/2026'),
  ('Camila Barbosa',   'Suporte',        'livre',   '');

-- Grupos de licenca
INSERT INTO grupos_licenca (nome, obs) VALUES
  ('Licenca Fisica 01', 'XE3 + D12 mesma chave'),
  ('Licenca Fisica 02', 'XE3 + D12 mesma chave'),
  ('Licenca Fisica 03', 'Upgrade D12 avulso'),
  ('Licenca Fisica 04', 'XE3 + D12 + Interbase');

-- Licencas vinculadas a devs
INSERT INTO licencas (dev_id, versao, serial, tipo, canal, hostname, edn_login, edn_senha, cad_efetuado, data_cad, obs, grupo_id) VALUES
  (1, 'D12',          'D12A-XXXX-7F3K-9PLM', 'Professional', 'EDN',     'PC-CARLOS',   'carlos.dev@empresa.com',   'Str0ngPass1',  true,  '2025-03-15', '', 1),
  (1, 'XE3',          'XE3B-YYYY-4H2J-8QRS', 'Professional', 'EDN',     'PC-CARLOS',   'carlos.dev@empresa.com',   'Str0ngPass1',  true,  '2024-01-10', '', 1),
  (2, 'D12',          'D12C-ZZZZ-1A5B-6TUV', 'Enterprise',   'EDN',     'PC-ANA',      'ana.dev@empresa.com',      'AnaDev2025',    true,  '2025-06-20', 'Licenca enterprise compartilhada', 2),
  (2, 'XE3',          'XE3D-WWWW-3C7D-2EFG', 'Enterprise',   'EDN',     'PC-ANA',      'ana.dev@empresa.com',      'AnaDev2025',    true,  '2024-02-14', '', 2),
  (3, 'D12',          'D12E-VVVV-9H1I-4JKL', 'Professional', 'Network', 'PC-BRUNO',    'bruno.dev@empresa.com',    'BrunoDev1',     true,  '2025-08-01', '', 3),
  (4, 'D12',          'D12F-UUUU-6M2N-7OPQ', 'Professional', 'EDN',     'PC-JULIANA',  'juliana.dev@empresa.com',  'Juliana456',    true,  '2025-04-22', 'Uso exclusivo SPED', NULL),
  (4, 'Interbase',    'IBG-TTTT-8R3S-5WXY',  'Professional', 'EDN',     'PC-JULIANA',  'juliana.dev@empresa.com',  'Juliana456',    true,  '2025-04-22', '', 4),
  (6, 'D12',          'D12H-SSSS-2U4V-1ABC', 'Enterprise',   'Network', 'PC-FERNANDA', 'fernanda.dev@empresa.com', 'FernDev1',     true,  '2025-09-10', '', NULL),
  (7, 'D12',          'D12I-RRRR-5W6X-3DEF', 'Professional', 'EDN',     'PC-PEDRO',    'pedro.dev@empresa.com',    'PedrOps1',     false, '2026-01-05', 'Aguardando cadastro EDN', NULL),
  (8, 'D12',          'D12J-QQQQ-7Y8Z-9GHI', 'Professional', 'EDN',     'PC-MARIANA',  'mariana.dev@empresa.com',  'Mari2026',      true,  '2025-11-18', '', NULL),
  (8, 'XE3',          'XE3K-PPPP-1A2B-4JKL', 'Professional', 'EDN',     'PC-MARIANA',  'mariana.dev@empresa.com',  'Mari2026',      true,  '2024-05-30', '', 4);

-- Licencas livres (sem dev)
INSERT INTO licencas (dev_id, versao, serial, tipo, canal, hostname, edn_login, edn_senha, cad_efetuado, data_cad, obs, grupo_id) VALUES
  (NULL, 'D12',          'D12L-OOOO-3C4D-6MNO', 'Professional', 'EDN',     '', '', '', false, NULL, 'Licenca reserva', NULL),
  (NULL, 'XE3',          'XE3M-NNNN-5E6F-8PQR', 'Enterprise',   'EDN',     '', '', '', false, NULL, 'Sem uso definido', NULL),
  (NULL, 'HTML5Builder', 'H5BN-MMMM-7G8H-2STU', 'Professional', 'EDN',     '', '', '', false, NULL, '', NULL),
  (NULL, 'Interbase',    'IBO-LLLL-9I1J-4VWX',  'Professional', 'Network', '', '', '', false, NULL, 'Disponivel para projeto novo', NULL);

-- Componentes
INSERT INTO componentes (nome, versao, ferramenta, informacoes, uso, licenciamento, site, obs, serial, usuario, senha) VALUES
  ('FastReport VCL',      '2024.1', 'D12', 'Gerador de relatorios com designer visual',           'Todos os modulos',       'Pago', 'https://www.fast-report.com',       '', 'FR-VCL-2024-XXXX-YYYY', 'empresa@fastreport.com', 'Frast2024'),
  ('TMS Component Pack',  '11.2',   'D12', 'Pacote com 600+ componentes VCL',                     'Phoenix Completa',       'Pago', 'https://www.tmssoftware.com',       '', 'TMS-PACK-1122-ZZZZ',    'dev@empresa.com',        'Tms2025'),
  ('DevExpress VCL',      '23.2',   'D12', 'Suite de componentes UI premium',                     'Interface administrativa','Pago', 'https://www.devexpress.com',        'Licenca site (ilimitada)', 'DX-VCL-2302-SITE-AAAA', 'admin@empresa.com', 'Dxsite23'),
  ('ACBR',                '1.3.2',  'D12', 'Automacao comercial: NFe, NFCe, CTe, MDFe, SAT',     'Fiscal e Contabil',      'Free', 'https://projetoacbr.com.br',        'Open source, comunidade ativa', '', '', ''),
  ('Indy',                '10.6',   'D12', 'Componentes de comunicacao TCP/IP, HTTP, SMTP, FTP',  'Todos os modulos',       'Free', 'https://www.indyproject.org',        'Incluido no IDE', '', '', ''),
  ('TMS FNC UI Pack',     '4.2',    'D12', 'Componentes cross-platform (VCL, FMX, Web, Mobile)',  'App mobile',             'Pago', 'https://www.tmssoftware.com',       '', 'TMS-FNC-0422-BBBB', 'dev@empresa.com', 'Tms2025'),
  ('madExcept',           '5.1',    'D12', 'Captura e relatorio de excecoes em runtime',          'Todos os modulos',       'Pago', 'https://madshi.net',                '', 'MAD-EXC-0510-CCCC', '', ''),
  ('EurekaLog',           '7.9',    'XE3', 'Tratamento avancado de excecoes e bug tracking',      'Modulos legados',        'Pago', 'https://www.eurekalog.com',         'Usado apenas em XE3', 'EL-PRO-0790-DDDD', 'eurekalog@empresa.com', 'Eureka7'),
  ('Spring4D',            '2.0',    'D12', 'Framework de injecao de dependencia e colecoes',      'Phoenix Completa',       'Free', 'https://spring4d.org',              '', '', '', ''),
  ('Boss',                '3.1',    'D12', 'Gerenciador de pacotes (similar ao npm)',              'Todos os modulos',       'Free', 'https://github.com/HashLoad/boss',  '', '', '', ''),
  ('Horse',               '3.0',    'D12', 'Framework REST API minimalista',                      'APIs internas',          'Free', 'https://github.com/HashLoad/horse', '', '', '', ''),
  ('Fortes Report CE',    '4.0',    'D12', 'Gerador de relatorios open source',                   'Relatorios simples',     'Free', 'https://github.com/fortesinformatica/fortesreport-ce', '', '', '', '');

-- Licencas Network
INSERT INTO licencas_network (sku, descricao, tipo, total_seats, login_name, senha, obs) VALUES
  ('RAD-D12-NET-PRO', 'RAD Studio D12 Professional Network Named', 'Professional', 5,  'acct-empresa-001', 'Netw0rkPro',  'License Certificate: LC-2025-001'),
  ('RAD-D12-NET-ENT', 'RAD Studio D12 Enterprise Network Named',   'Enterprise',   3,  'acct-empresa-002', 'Netw0rkEnt',  'License Certificate: LC-2025-002'),
  ('IB-2020-NET',     'InterBase 2020 Server Network',             'Professional', 10, 'acct-empresa-003', 'IbNet2020',   'Servidor: srv-interbase.local');
