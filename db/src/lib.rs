pub struct Db {

}

impl Db {
    pub async fn get(&self) {

    }
}


#[cfg(test)]
mod tests {
    #[test]
    fn it_works() {
        let result = 2 + 2;
        assert_eq!(result, 4);
    }
}



#[tokio::async_trait]
pub trait Database {
    pub async fn get(&self, )
}