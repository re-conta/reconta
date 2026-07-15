Permita compartilhar suas transações e relatórios com outros usuários cadastrados no site, o usuário poderá selecionar quais periodos e quais contas bancárias deseja compartilhar, deve haver uma opção para habilitar ou desabilitar o compartilhamento de transações futuras, permita tambem que a pessoa que compartilha escolha se o usuário convidado poderá editar as transações ou não. A pessoa que recebe o convite deve receber uma notificação mas tome cuidado, essa notificação não pode ser marcada como lida e não ser que o usuário convidado rejeite, o usuário que convidou deve ter a opção de cancelar o compartilhamento a qualquer momento. Quando o usuário aceita ou rejeita quem compartilhou deve ser notificado. Lembre-se, todas as notificações devem ser em tempo real(sem necessidade de refresh).

Em desktops algumas páginas estão muito estreitas(nas rotas /tags, /categorias entre outras), preciso que em desktops deixe o layout(parte central onde é mostrado o conteúdo) um pouco mais largo, em celulares eu preciso da mesma padronização porem preciso que o layout ocupe a maior área horizontal para otimizar a visualização nestes dispositivos de tela pequena. Use seu bom gosto e siga os padrões de cores e fontes do site. 

Na página inicial durante o carregamento onde há escrito: "ReConta seu dinheiro com clareza" primeiro o texto carrega "expremido" e depois ele se ajusta, ache uma forma criativa de evitar esse tipo de "layout shift"

Em /relatorios sob "Tudo" e "Intervalo personalizado" eu preciso do fluxo tambem, preciso de um gráfico por linha e não um ao lado do outro.

Crie uma página administrativa seguindo o layout do site, o usuário com o e-mail sistematico@gmail.com ou lsbrum@icloud.com ou reconta@reconta.app será sempre o Super Admin, os cargos do site são: 1- Pessoa Física(padrão), 2- Pessoa Jurídica(será necessário no cadastro um CNPJ válido), 3- Contador ou Técnico Contábil, 4- Administrador, 5- Super Administrador(eu), apenas Super Administrador e Administrador tem acesso ao painel de admin, mas as permissões devem poder serem editadas.

No painel de admin crie uma página de estatísticas com um set completo de estatíticas: Visitas únicas, Visitas, quais páginas visitadas, ips, agentes de navegador, gráficos(com a opção de selecionar o range), referrer localização(por ip) usando IP real(fornecido pela cloudflare e passada para o nginx) e o sistema de GeoIP2 Lite da Maxmind, para isso preciso que ajuste na VPS esse sistema e alem disso crie este painel com mais informações que achar pertinentes, documente tudo no README.md

No painel de admin crie uma página de logs onde toda visita deve ser logada assim como agente, ip, navegador, sistema e as páginas que visitou e as ações que este usuário fez no site.

Transforme a pasta files/ em ansible/ crie um playbook simples mais completo usando o ansible com:

- Suporte as units systemd do go
- Proxy reverso para o vue e go usando o nginx
- Instalação de certificados letsencrypt(caso eles não existam, cuidado com o block por parte do cloudflare/letsencrypt)
- Instalação do go, nginx, bun, node(se ainda não estiverem instalados), mas fique atento o bun na vps é instalado na home do usuário nginx
- Criação dos usuários e grupos caso ainda não existam(nginx por exemplo)
