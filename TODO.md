Transforme a pasta files/ em ansible/ crie um playbook simples mais completo usando o ansible com:

- Suporte as units systemd do go
- Proxy reverso para o vue e go usando o nginx
- Instalação de certificados letsencrypt(caso eles não existam, cuidado com o block por parte do cloudflare/letsencrypt)
- Instalação do go, nginx, bun, node(se ainda não estiverem instalados), mas fique atento o bun na vps é instalado na home do usuário nginx
- Criação dos usuários e grupos caso ainda não existam(nginx por exemplo)