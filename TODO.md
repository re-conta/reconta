Usando o plano gratuito do Groq(https://groq.com) crie um sistema que analisa a conta do usuário com base na "saúde" oferece recomendações de meios de economizar ou de investir o dinheiro que sobra(no caso de saúde >= 3), para dicas de corte de gastos ou dicas de investimentos analize as transações, como por exemplo: 

- + de 1 Streaming recomende cancelar os excedentes deixando um apenas
- Gastos com combustível(procure se na região do usuários existem postos com combustível mais barato)
- Não se prenda a esses dois eventos e seja criativo(usando a api do Groq para "pensar" em tempo real e recomendar).

Na consulta usando o Groq tome muito cuidado para não exceder os limites da API gratuita, ou seja, as interações devem obedecer uma "fila" para executar apenas uma requisição por vez e inserir um delay/debounce/throttle indepente do número de usuários que estejam sendo processados para não exceder os limites horários, diários, semanais e mensais em hipótese alguma.

Em /relatorios sob "Tudo" e "Intervalo personalizado" eu preciso do fluxo tambem, preciso de um gráfico por linha e não um ao lado do outro.

No painel de admin crie uma página de logs onde toda visita deve ser logada assim como agente, ip, navegador, sistema e as páginas que visitou e as ações que este usuário fez no site.

Transforme a pasta files/ em ansible/ crie um playbook simples mais completo usando o ansible com:

- Suporte as units systemd do go
- Proxy reverso para o vue e go usando o nginx
- Instalação de certificados letsencrypt(caso eles não existam, cuidado com o block por parte do cloudflare/letsencrypt)
- Instalação do go, nginx, bun, node(se ainda não estiverem instalados), mas fique atento o bun na vps é instalado na home do usuário nginx
- Criação dos usuários e grupos caso ainda não existam(nginx por exemplo)
